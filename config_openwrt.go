//go:build openwrt

// Copyright (c) 2025 Kilimcinin Kör Oğlu <k@keremgok.tr>
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type AppConfig struct {
	Driver              string `json:"driver"`
	Device              string `json:"device"`
	Slot                int    `json:"slot"`
	Timeout             int    `json:"timeout"`
	EnableOutputLogs    bool   `json:"enable_output_logs"`
	AutoNotification    bool   `json:"auto_notification"`
	RebootMethod        string `json:"reboot_method"`
	RebootATCommand     string `json:"reboot_at_command"`
	RebootATDevice      string `json:"reboot_at_device"`
	RebootQMIDevice     string `json:"reboot_qmi_device"`
	RebootQMISlot       int    `json:"reboot_qmi_slot"`
	RebootMBIMDevice    string `json:"reboot_mbim_device"`
	RebootCustomCommand string `json:"reboot_custom_command"`
}

func loadConfig() (*AppConfig, error) {
	config := &AppConfig{
		Driver:        "auto",
		Slot:          1,
		Timeout:       30,
		RebootMethod:  "none",
		RebootQMISlot: 1,
	}

	if v := uciGet("hermes_euicc.config.driver"); v != "" {
		config.Driver = v
	}

	switch config.Driver {
	case "qmi":
		config.Device = uciGet("hermes_euicc.config.qmi_device")
	case "mbim":
		config.Device = uciGet("hermes_euicc.config.mbim_device")
	case "at":
		config.Device = uciGet("hermes_euicc.config.at_device")
	default:
		config.Device = uciGet("hermes_euicc.config.qmi_device")
		if config.Device == "" {
			config.Device = uciGet("hermes_euicc.config.mbim_device")
		}
		if config.Device == "" {
			config.Device = uciGet("hermes_euicc.config.at_device")
		}
	}

	if v := uciGet("hermes_euicc.config.qmi_sim_slot"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			config.Slot = n
		}
	}
	if v := uciGet("hermes_euicc.config.timeout"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			config.Timeout = n
		}
	}

	config.EnableOutputLogs = uciGet("hermes_euicc.config.enable_output_logs") == "1"
	config.AutoNotification = uciGet("hermes_euicc.config.auto_notification") == "1"

	if v := uciGet("hermes_euicc.config.reboot_method"); v != "" {
		config.RebootMethod = v
	}
	config.RebootATCommand = uciGet("hermes_euicc.config.reboot_at_command")
	config.RebootATDevice = uciGet("hermes_euicc.config.reboot_at_device")
	config.RebootQMIDevice = uciGet("hermes_euicc.config.reboot_qmi_device")
	if v := uciGet("hermes_euicc.config.reboot_qmi_slot"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			config.RebootQMISlot = n
		}
	}
	config.RebootMBIMDevice = uciGet("hermes_euicc.config.reboot_mbim_device")
	config.RebootCustomCommand = uciGet("hermes_euicc.config.reboot_custom_command")

	return config, nil
}

func saveConfig(config *AppConfig) error {
	uciSet("hermes_euicc.config.driver", config.Driver)

	switch config.Driver {
	case "qmi":
		uciSet("hermes_euicc.config.qmi_device", config.Device)
	case "mbim":
		uciSet("hermes_euicc.config.mbim_device", config.Device)
	case "at":
		uciSet("hermes_euicc.config.at_device", config.Device)
	}

	uciSet("hermes_euicc.config.qmi_sim_slot", fmt.Sprintf("%d", config.Slot))
	uciSet("hermes_euicc.config.timeout", fmt.Sprintf("%d", config.Timeout))
	uciSetBool("hermes_euicc.config.enable_output_logs", config.EnableOutputLogs)
	uciSetBool("hermes_euicc.config.auto_notification", config.AutoNotification)
	uciSet("hermes_euicc.config.reboot_method", config.RebootMethod)
	uciSet("hermes_euicc.config.reboot_at_command", config.RebootATCommand)
	uciSet("hermes_euicc.config.reboot_at_device", config.RebootATDevice)
	uciSet("hermes_euicc.config.reboot_qmi_device", config.RebootQMIDevice)
	uciSet("hermes_euicc.config.reboot_qmi_slot", fmt.Sprintf("%d", config.RebootQMISlot))
	uciSet("hermes_euicc.config.reboot_mbim_device", config.RebootMBIMDevice)
	uciSet("hermes_euicc.config.reboot_custom_command", config.RebootCustomCommand)

	return uciCommit()
}

func uciGet(key string) string {
	out, err := exec.Command("uci", "get", key).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func uciSet(key, value string) {
	exec.Command("uci", "set", key+"="+value).Run()
}

func uciSetBool(key string, value bool) {
	if value {
		uciSet(key, "1")
	} else {
		uciSet(key, "0")
	}
}

func uciCommit() error {
	return exec.Command("uci", "commit", "hermes_euicc").Run()
}
