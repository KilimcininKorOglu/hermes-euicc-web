// Copyright (c) 2025 Kilimcinin Kör Oğlu <k@keremgok.tr>
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

func execReboot(config *AppConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	switch config.RebootMethod {
	case "at":
		return rebootAT(ctx, config)
	case "qmi":
		return rebootQMI(ctx, config)
	case "mbim":
		return rebootMBIM(ctx, config)
	case "custom":
		return rebootCustom(ctx, config)
	case "none", "":
		return fmt.Errorf("no reboot method configured")
	default:
		return fmt.Errorf("unknown reboot method: %s", config.RebootMethod)
	}
}

func rebootAT(ctx context.Context, config *AppConfig) error {
	atCmd := config.RebootATCommand
	if atCmd == "" {
		atCmd = "AT+CFUN=1,1"
	}
	device := config.RebootATDevice
	if device == "" {
		return fmt.Errorf("AT reboot device not configured")
	}

	cmd := exec.CommandContext(ctx, "sh", "-c",
		fmt.Sprintf("echo '%s' > %s", atCmd, device))
	return cmd.Run()
}

func rebootQMI(ctx context.Context, config *AppConfig) error {
	device := config.RebootQMIDevice
	if device == "" {
		device = "/dev/cdc-wdm0"
	}

	cmd := exec.CommandContext(ctx, "uqmi", "-d", device,
		"--set-device-operating-mode", "reset")
	return cmd.Run()
}

func rebootMBIM(ctx context.Context, config *AppConfig) error {
	device := config.RebootMBIMDevice
	if device == "" {
		device = "/dev/cdc-wdm0"
	}

	off := exec.CommandContext(ctx, "mbimcli", "-d", device,
		"--set-radio-state=off")
	if err := off.Run(); err != nil {
		return fmt.Errorf("mbim radio off: %w", err)
	}

	time.Sleep(1 * time.Second)

	on := exec.CommandContext(ctx, "mbimcli", "-d", device,
		"--set-radio-state=on")
	if err := on.Run(); err != nil {
		return fmt.Errorf("mbim radio on: %w", err)
	}

	return nil
}

func rebootCustom(ctx context.Context, config *AppConfig) error {
	if config.RebootCustomCommand == "" {
		return fmt.Errorf("custom reboot command not configured")
	}
	cmd := exec.CommandContext(ctx, "sh", "-c", config.RebootCustomCommand)
	return cmd.Run()
}
