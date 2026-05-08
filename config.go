//go:build !openwrt

// Copyright (c) 2025 Kilimcinin Kör Oğlu <k@keremgok.tr>
// SPDX-License-Identifier: MIT

package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type AppConfig struct {
	Driver             string `json:"driver"`
	Device             string `json:"device"`
	Slot               int    `json:"slot"`
	Timeout            int    `json:"timeout"`
	EnableOutputLogs   bool   `json:"enable_output_logs"`
	AutoNotification   bool   `json:"auto_notification"`
	RebootMethod       string `json:"reboot_method"`
	RebootATCommand    string `json:"reboot_at_command"`
	RebootATDevice     string `json:"reboot_at_device"`
	RebootQMIDevice    string `json:"reboot_qmi_device"`
	RebootQMISlot      int    `json:"reboot_qmi_slot"`
	RebootMBIMDevice   string `json:"reboot_mbim_device"`
	RebootCustomCommand string `json:"reboot_custom_command"`
}

func loadConfig() (*AppConfig, error) {
	config := &AppConfig{
		Driver:       "auto",
		Slot:         1,
		Timeout:      30,
		RebootMethod: "none",
		RebootQMISlot: 1,
	}

	path := findConfigPath()
	if path == "" {
		return config, nil
	}

	file, err := os.Open(path)
	if err != nil {
		return config, nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), "\"'")

		switch key {
		case "driver":
			if value != "" {
				config.Driver = value
			}
		case "device":
			config.Device = value
		case "slot":
			if n, err := strconv.Atoi(value); err == nil && n > 0 {
				config.Slot = n
			}
		case "timeout":
			if n, err := strconv.Atoi(value); err == nil && n > 0 {
				config.Timeout = n
			}
		case "enable_output_logs":
			config.EnableOutputLogs = value == "1" || value == "true"
		case "auto_notification":
			config.AutoNotification = value == "1" || value == "true"
		case "reboot_method":
			config.RebootMethod = value
		case "reboot_at_command":
			config.RebootATCommand = value
		case "reboot_at_device":
			config.RebootATDevice = value
		case "reboot_qmi_device":
			config.RebootQMIDevice = value
		case "reboot_qmi_slot":
			if n, err := strconv.Atoi(value); err == nil && n > 0 {
				config.RebootQMISlot = n
			}
		case "reboot_mbim_device":
			config.RebootMBIMDevice = value
		case "reboot_custom_command":
			config.RebootCustomCommand = value
		}
	}
	return config, scanner.Err()
}

func saveConfig(config *AppConfig) error {
	path := findConfigPath()
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("cannot determine home directory: %w", err)
		}
		dir := filepath.Join(home, ".config", "hermes-euicc")
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("cannot create config directory: %w", err)
		}
		path = filepath.Join(dir, "config")
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("driver=%s\n", config.Driver))
	if config.Device != "" {
		sb.WriteString(fmt.Sprintf("device=%s\n", config.Device))
	}
	sb.WriteString(fmt.Sprintf("slot=%d\n", config.Slot))
	sb.WriteString(fmt.Sprintf("timeout=%d\n", config.Timeout))
	if config.EnableOutputLogs {
		sb.WriteString("enable_output_logs=1\n")
	}
	if config.AutoNotification {
		sb.WriteString("auto_notification=1\n")
	}
	if config.RebootMethod != "" && config.RebootMethod != "none" {
		sb.WriteString(fmt.Sprintf("reboot_method=%s\n", config.RebootMethod))
		switch config.RebootMethod {
		case "at":
			if config.RebootATCommand != "" {
				sb.WriteString(fmt.Sprintf("reboot_at_command=%s\n", config.RebootATCommand))
			}
			if config.RebootATDevice != "" {
				sb.WriteString(fmt.Sprintf("reboot_at_device=%s\n", config.RebootATDevice))
			}
		case "qmi":
			if config.RebootQMIDevice != "" {
				sb.WriteString(fmt.Sprintf("reboot_qmi_device=%s\n", config.RebootQMIDevice))
			}
			sb.WriteString(fmt.Sprintf("reboot_qmi_slot=%d\n", config.RebootQMISlot))
		case "mbim":
			if config.RebootMBIMDevice != "" {
				sb.WriteString(fmt.Sprintf("reboot_mbim_device=%s\n", config.RebootMBIMDevice))
			}
		case "custom":
			if config.RebootCustomCommand != "" {
				sb.WriteString(fmt.Sprintf("reboot_custom_command=%s\n", config.RebootCustomCommand))
			}
		}
	}

	return os.WriteFile(path, []byte(sb.String()), 0644)
}

func findConfigPath() string {
	if _, err := os.Stat("hermes-euicc.conf"); err == nil {
		return "hermes-euicc.conf"
	}
	if home, err := os.UserHomeDir(); err == nil {
		p := filepath.Join(home, ".config", "hermes-euicc", "config")
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}
