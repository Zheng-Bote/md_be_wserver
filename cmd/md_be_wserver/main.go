/**
 * SPDX-FileComment: Medical Devices Backend
 * SPDX-FileType: SOURCE
 * SPDX-FileContributor: ZHENG Robert
 * SPDX-FileCopyrightText: 2026 ZHENG Robert
 * SPDX-License-Identifier: Apache-2.0

 * @file main.go
 * @brief Main entry point for the Webserver service
 * @version 0.1.0
 * @date 2026-08-16
 *
 * @author ZHENG Robert (robert@hase-zheng.net)
 * @copyright Copyright (c) 2026 ZHENG Robert
 * @LICENSE Apache-2.0
 */

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"md_be_wserver/internal/config"
	"md_be_wserver/internal/db"
	"md_be_wserver/internal/http"
)

var (
	appName        = "MitM Medical Devices Backend"
	appDescription = "Backend server for Medical Devices"
	version        = "1.0.0"
)

func main() {
	var configPath string
	if len(os.Args) < 2 {
		execPath, err := os.Executable()
		if err == nil {
			execDir := filepath.Dir(execPath)
			// check common config locations
			possiblePaths := []string{
				filepath.Join(execDir, "md_be_wserver_config.json"),
				filepath.Join(execDir, "md_be_wserver_config.enc"),
				"../data/md_be_wserver_config.json",
				filepath.Join(execDir, "data", "md_be_wserver_config.json"),
			}
			for _, p := range possiblePaths {
				if _, err := os.Stat(p); err == nil {
					configPath = p
					break
				}
			}
		}
		if configPath == "" {
			fmt.Println("Usage: md_be_wserver <path/to/md_be_wserver_config.json>")
			os.Exit(1)
		}
	} else {
		configPath = os.Args[1]
	}

	password := os.Getenv("MD_BE_WSERVER_PASSWORD")
	if password == "" {
		log.Println("ERROR: MD_BE_WSERVER_PASSWORD is not set")
		os.Exit(1)
	}

	cfg, err := config.LoadConfig(configPath, password)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	repo, err := db.NewRepository(cfg.DB)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer repo.Close()
	log.Println("Database connection established")

	// Sync Configured Admins to DB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := repo.SyncAdmins(ctx, cfg.Admins); err != nil {
		log.Fatalf("Failed to sync admin users: %v", err)
	}

	httpServer := &http.Server{
		Port:           cfg.WServer.HTTPPort,
		UseHTTPS:       cfg.WServer.UseHTTPS,
		SSLCert:        cfg.WServer.SSLCert,
		SSLKey:         cfg.WServer.SSLKey,
		Admins:         cfg.Admins,
		AppName:        appName,
		AppDescription: appDescription,
		AppVersion:     version,
		Config:         cfg.WServer,
		Repo:           repo,
	}

	if err := httpServer.Start(); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}

	protocol := "HTTP"
	if cfg.WServer.UseHTTPS {
		protocol = "HTTPS"
	}
	log.Printf("%s Server listening on port %d", protocol, cfg.WServer.HTTPPort)

	// Wait for termination
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := httpServer.Stop(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}
}
