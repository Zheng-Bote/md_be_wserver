/**
 * SPDX-FileComment: Medical Devices Backend
 * SPDX-FileType: SOURCE
 * SPDX-FileContributor: ZHENG Robert
 * SPDX-FileCopyrightText: 2026 ZHENG Robert
 * SPDX-License-Identifier: Apache-2.0
 *
 * @file handlers_auth.go
 * @brief Authentication HTTP handlers
 * @version 1.0.0
 * @date 2026-08-16
 *
 * @author ZHENG Robert (robert@hase-zheng.net)
 * @copyright Copyright (c) 2026 ZHENG Robert
 * @LICENSE Apache-2.0
 */

package http

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"image/png"
	"net/http"
	"strings"
	"time"

	"github.com/pquerna/otp/totp"

	"md_be_wserver/internal/crypto"
)

type claimsKey struct{}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	TotpCode string `json:"totp_code,omitempty"`
}

type LoginResponse struct {
	AccessToken   string `json:"access_token"`
	RefreshToken  string `json:"refresh_token"`
	ExpiresIn     int    `json:"expires_in"`
	MustChangePwd bool   `json:"must_change_pwd"`
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := s.Repo.GetUserByLoginID(r.Context(), req.Username)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if !user.IsActive {
		http.Error(w, "Account disabled", http.StatusForbidden)
		return
	}

	if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
		http.Error(w, "Account locked temporarily", http.StatusTooManyRequests)
		return
	}

	valid, err := crypto.VerifyPassword(req.Password, user.PasswordHash)
	if err != nil || !valid {
		attempts := user.FailedLoginAttempts + 1
		var lockedUntil *time.Time
		if attempts >= 5 {
			lockTime := time.Now().Add(15 * time.Minute)
			lockedUntil = &lockTime
		}
		_ = s.Repo.UpdateFailedLogins(r.Context(), user.ID, attempts, lockedUntil)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if user.FailedLoginAttempts > 0 {
		_ = s.Repo.ResetFailedLogins(r.Context(), user.ID)
	}

	mustSetupTOTP := false
	if s.Config.RequireTOTP && !user.TotpEnabled {
		mustSetupTOTP = true
	} else if user.TotpEnabled {
		if req.TotpCode == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusPreconditionRequired) // 428
			w.Write([]byte(`{"error": "totp_required"}`))
			return
		}
		if user.TotpSecret == nil || !totp.Validate(req.TotpCode, *user.TotpSecret) {
			http.Error(w, "Invalid TOTP code", http.StatusUnauthorized)
			return
		}
	}

	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	sessionID, err := s.Repo.CreateSession(r.Context(), user.ID, r.RemoteAddr, r.UserAgent(), expiresAt)
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	accessToken, err := GenerateJWT(s.Config.JWTSecret, user.ID, user.LoginID, user.MustChangePwd, mustSetupTOTP)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	resp := LoginResponse{
		AccessToken:   accessToken,
		RefreshToken:  sessionID,
		ExpiresIn:     900,
		MustChangePwd: user.MustChangePwd,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := ValidateJWT(s.Config.JWTSecret, tokenString)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if claims.MustChangePwd && r.URL.Path != "/api/v1/auth/change-password" {
			http.Error(w, "Password change required", http.StatusForbidden)
			return
		}

		if claims.MustSetupTOTP && r.URL.Path != "/api/v1/auth/totp/setup" && r.URL.Path != "/api/v1/auth/totp/verify" && r.URL.Path != "/api/v1/auth/logout" {
			http.Error(w, "TOTP setup required", http.StatusForbidden)
			return
		}

		ctx := context.WithValue(r.Context(), claimsKey{}, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (s *Server) handleRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID, err := s.Repo.GetSession(r.Context(), req.RefreshToken)
	if err != nil {
		http.Error(w, "Invalid or expired session", http.StatusUnauthorized)
		return
	}

	var loginID string
	var mustChangePwd bool
	var isActive bool
	var totpEnabled bool
	err = s.Repo.Pool.QueryRow(r.Context(), "SELECT login_id, must_change_pwd, is_active, totp_enabled FROM users WHERE id = $1", userID).Scan(&loginID, &mustChangePwd, &isActive, &totpEnabled)
	if err != nil || !isActive {
		http.Error(w, "Invalid or disabled user", http.StatusUnauthorized)
		return
	}
	
	mustSetupTOTP := false
	if s.Config.RequireTOTP && !totpEnabled {
		mustSetupTOTP = true
	}

	accessToken, err := GenerateJWT(s.Config.JWTSecret, userID, loginID, mustChangePwd, mustSetupTOTP)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	resp := LoginResponse{
		AccessToken:   accessToken,
		RefreshToken:  req.RefreshToken,
		ExpiresIn:     900,
		MustChangePwd: mustChangePwd,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}
	
	_ = s.Repo.RevokeSession(r.Context(), req.RefreshToken)
	w.WriteHeader(http.StatusOK)
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

func (s *Server) handleChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims := r.Context().Value(claimsKey{}).(*Claims)
	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := s.Repo.GetUserByID(r.Context(), claims.UserID)
	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	valid, err := crypto.VerifyPassword(req.OldPassword, user.PasswordHash)
	if err != nil || !valid {
		http.Error(w, "Invalid old password", http.StatusUnauthorized)
		return
	}

	newHash, err := crypto.HashPassword(req.NewPassword)
	if err != nil {
		http.Error(w, "Failed to process new password", http.StatusInternalServerError)
		return
	}

	if err := s.Repo.UpdatePassword(r.Context(), user.ID, newHash); err != nil {
		http.Error(w, "Failed to update password", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Password updated successfully"}`))
}

func (s *Server) handleTOTPSetup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims := r.Context().Value(claimsKey{}).(*Claims)
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      s.AppName,
		AccountName: claims.LoginID,
	})
	if err != nil {
		http.Error(w, "Failed to generate TOTP", http.StatusInternalServerError)
		return
	}

	err = s.Repo.UpdateTOTPSecret(r.Context(), claims.UserID, key.Secret())
	if err != nil {
		http.Error(w, "Failed to save TOTP secret", http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	img, _ := key.Image(200, 200)
	_ = png.Encode(&buf, img)
	qrBase64 := base64.StdEncoding.EncodeToString(buf.Bytes())

	resp := map[string]string{
		"secret":  key.Secret(),
		"qr_code": "data:image/png;base64," + qrBase64,
		"uri":     key.URL(),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleTOTPVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims := r.Context().Value(claimsKey{}).(*Claims)
	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	user, err := s.Repo.GetUserByID(r.Context(), claims.UserID)
	if err != nil || user.TotpSecret == nil {
		http.Error(w, "TOTP not setup", http.StatusBadRequest)
		return
	}

	valid := totp.Validate(req.Code, *user.TotpSecret)
	if !valid {
		http.Error(w, "Invalid TOTP code", http.StatusUnauthorized)
		return
	}

	if err := s.Repo.EnableTOTP(r.Context(), claims.UserID); err != nil {
		http.Error(w, "Failed to enable TOTP", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "TOTP successfully verified and enabled"}`))
}
