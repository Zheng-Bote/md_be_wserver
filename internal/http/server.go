/**
 * SPDX-FileComment: Medical Devices Backend
 * SPDX-FileType: SOURCE
 * SPDX-FileContributor: ZHENG Robert
 * SPDX-FileCopyrightText: 2026 ZHENG Robert
 * SPDX-License-Identifier: Apache-2.0
 *
 * @file server.go
 * @brief HTTP API server with health, info, and time endpoints
 * @version 0.1.0
 * @date 2026-08-16
 *
 * @author ZHENG Robert (robert@hase-zheng.net)
 * @copyright Copyright (c) 2026 ZHENG Robert
 * @LICENSE Apache-2.0
 */

package http

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/google/uuid"

	"md_be_wserver/internal/config"
	"md_be_wserver/internal/db"
)

type requestIDKey struct{}

func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic recovered: %v\n%s", err, debug.Stack())
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func LimitBody(next http.Handler, maxBodyBytes int64) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
		next.ServeHTTP(w, r)
	})
}

func WithRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = uuid.NewString()
		}
		ctx := context.WithValue(r.Context(), requestIDKey{}, id)
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func Throttle(max int) func(http.Handler) http.Handler {
	sem := make(chan struct{}, max)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
				next.ServeHTTP(w, r)
			default:
				http.Error(w, "server busy", http.StatusServiceUnavailable)
			}
		})
	}
}

type Server struct {
	Port           int
	UseHTTPS       bool
	SSLCert        string
	SSLKey         string
	Admins         []config.AdminUser
	AppName        string
	AppDescription string
	AppVersion     string
	Config         config.WebServerConfig
	Repo           *db.Repository
	httpSrv        *http.Server
}

func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			s.handleInfo(w, r)
			return
		}
		http.NotFound(w, r)
	})
	mux.HandleFunc("/info", s.handleInfo)
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/time", s.handleTime)

	// Auth Endpoints
	mux.HandleFunc("/api/v1/auth/login", s.handleLogin)
	mux.HandleFunc("/api/v1/auth/refresh", s.handleRefresh)
	
	// Protected Endpoints
	mux.Handle("/api/v1/auth/logout", s.RequireAuth(http.HandlerFunc(s.handleLogout)))
	mux.Handle("/api/v1/auth/change-password", s.RequireAuth(http.HandlerFunc(s.handleChangePassword)))
	mux.Handle("/api/v1/auth/totp/setup", s.RequireAuth(http.HandlerFunc(s.handleTOTPSetup)))
	mux.Handle("/api/v1/auth/totp/verify", s.RequireAuth(http.HandlerFunc(s.handleTOTPVerify)))

	// Devices Endpoints (Protected)
	mux.Handle("/api/v1/devices/get_device_types", s.RequireAuth(http.HandlerFunc(s.handleGetDeviceTypes)))
	mux.Handle("/api/v1/devices/get_devices", s.RequireAuth(http.HandlerFunc(s.handleGetDevices)))

	var handler http.Handler = mux

	limit := int64(1 << 20) // default 1MB
	if s.Config.LimitBody > 0 {
		limit = s.Config.LimitBody
	}
	handler = LimitBody(handler, limit)

	handler = WithRequestID(handler)

	throttleMax := 100
	if s.Config.Throttle > 0 {
		throttleMax = s.Config.Throttle
	}
	handler = Throttle(throttleMax)(handler)

	handler = Recover(handler)

	readTimeout := 5 * time.Second
	if s.Config.ReadTimeout > 0 {
		readTimeout = time.Duration(s.Config.ReadTimeout) * time.Second
	}
	readHeaderTimeout := 3 * time.Second
	if s.Config.ReadHeaderTimeout > 0 {
		readHeaderTimeout = time.Duration(s.Config.ReadHeaderTimeout) * time.Second
	}
	writeTimeout := 30 * time.Second
	if s.Config.WriteTimeout > 0 {
		writeTimeout = time.Duration(s.Config.WriteTimeout) * time.Second
	}
	idleTimeout := 120 * time.Second
	if s.Config.IdleTimeout > 0 {
		idleTimeout = time.Duration(s.Config.IdleTimeout) * time.Second
	}

	s.httpSrv = &http.Server{
		Addr:              fmt.Sprintf(":%d", s.Port),
		Handler:           handler,
		ReadTimeout:       readTimeout,
		ReadHeaderTimeout: readHeaderTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}

	go func() {
		if s.UseHTTPS {
			if s.SSLCert == "" {
				s.SSLCert = "server.crt"
			}
			if s.SSLKey == "" {
				s.SSLKey = "server.key"
			}
			if err := s.httpSrv.ListenAndServeTLS(s.SSLCert, s.SSLKey); err != nil && err != http.ErrServerClosed {
				log.Printf("ERROR: Failed to start HTTPS server (cert: %s, key: %s): %v\n", s.SSLCert, s.SSLKey, err)
			}
		} else {
			if err := s.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Printf("ERROR: Failed to start HTTP server: %v\n", err)
			}
		}
	}()

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	if s.httpSrv != nil {
		return s.httpSrv.Shutdown(ctx)
	}
	return nil
}

func (s *Server) handleInfo(w http.ResponseWriter, r *http.Request) {
	info := map[string]string{
		"name":        s.AppName,
		"description": s.AppDescription,
		"version":     s.AppVersion,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

func (s *Server) handleTime(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	res := map[string]interface{}{
		"local_time": now.Format(time.RFC3339),
		"timestamp":  now.Unix(),
		"timezone":   now.Location().String(),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (s *Server) authenticate(r *http.Request) (string, bool) {
	user, token, ok := r.BasicAuth()
	if !ok {
		return "", false
	}
	for _, admin := range s.Admins {
		if admin.Username == user && admin.Password == token {
			return user, true
		}
	}
	return "", false
}
