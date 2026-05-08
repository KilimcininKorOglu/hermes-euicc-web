// Copyright (c) 2025 Kilimcinin Kör Oğlu <k@keremgok.tr>
// SPDX-License-Identifier: MIT

package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var version = "0.1.0"

func main() {
	port := flag.Int("port", 9090, "HTTP server port")
	bind := flag.String("bind", "0.0.0.0", "Bind address")
	hermesBinary := flag.String("hermes-binary", "hermes-euicc", "Path to hermes-euicc binary")
	hermesDriver := flag.String("hermes-driver", "", "Driver override (qmi/mbim/at/ccid)")
	hermesDevice := flag.String("hermes-device", "", "Device path override")
	hermesSlot := flag.Int("hermes-slot", 0, "SIM slot override")
	hermesTimeout := flag.Int("hermes-timeout", 30, "CLI command timeout in seconds")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	showVersion := flag.Bool("version", false, "Show version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("hermes-euicc-web %s\n", version)
		os.Exit(0)
	}

	logLevel := slog.LevelInfo
	if *verbose {
		logLevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel}))
	slog.SetDefault(logger)

	hermesConfig := &HermesConfig{
		Binary:  *hermesBinary,
		Driver:  *hermesDriver,
		Device:  *hermesDevice,
		Slot:    *hermesSlot,
		Timeout: *hermesTimeout,
	}

	i18n, err := NewI18n()
	if err != nil {
		slog.Error("failed to load translations", "error", err)
		os.Exit(1)
	}

	hermesClient := NewHermesClient(hermesConfig)
	srv := NewServer(i18n, hermesClient)
	addr := fmt.Sprintf("%s:%d", *bind, *port)

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		slog.Info("shutting down")
		os.Exit(0)
	}()

	slog.Info("starting hermes-euicc-web", "address", addr, "version", version)
	if err := http.ListenAndServe(addr, srv); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
