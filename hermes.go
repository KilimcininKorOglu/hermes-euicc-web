// Copyright (c) 2025 Kilimcinin Kör Oğlu <k@keremgok.tr>
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os/exec"
	"sync"
	"time"
)

type HermesClient struct {
	config *HermesConfig
	mu     sync.Mutex
}

type HermesResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   string          `json:"error,omitempty"`
}

type ChipInfoData struct {
	EID                string              `json:"eid"`
	ConfiguredAddresses *ConfiguredAddresses `json:"configured_addresses,omitempty"`
	EUICCInfo2         *EUICCInfo2          `json:"euicc_info2,omitempty"`
}

type ConfiguredAddresses struct {
	DefaultSMDPAddress string `json:"default_smdp_address"`
	RootSMDSAddress    string `json:"root_smds_address"`
}

type EUICCInfo2 struct {
	ProfileVersion       string            `json:"profile_version"`
	SVN                  string            `json:"svn"`
	EUICCFirmwareVer     string            `json:"euicc_firmware_ver"`
	GlobalPlatformVersion string           `json:"global_platform_version"`
	PPVersion            string            `json:"pp_version"`
	ExtCardResource      *ExtCardResource  `json:"ext_card_resource,omitempty"`
	SASAccreditationNumber string          `json:"sas_accreditation_number"`
}

type ExtCardResource struct {
	InstalledApplication  int `json:"installed_application"`
	FreeNonVolatileMemory int `json:"free_non_volatile_memory"`
	FreeVolatileMemory    int `json:"free_volatile_memory"`
}

type ProfileData struct {
	ICCID               string `json:"iccid"`
	ISDPAID             string `json:"isdp_aid"`
	ProfileState        int    `json:"profile_state"`
	ProfileName         string `json:"profile_name"`
	ProfileNickname     string `json:"profile_nickname"`
	ServiceProviderName string `json:"service_provider_name"`
	ProfileClass        string `json:"profile_class"`
}

type ActionResult struct {
	Message string `json:"message"`
	ICCID   string `json:"iccid,omitempty"`
}

type NotificationData struct {
	SequenceNumber             int    `json:"sequence_number"`
	ProfileManagementOperation int    `json:"profile_management_operation"`
	Address                    string `json:"address"`
	ICCID                      string `json:"iccid"`
}

type AutoNotificationResult struct {
	Message       string                    `json:"message"`
	Total         int                       `json:"total"`
	Processed     int                       `json:"processed"`
	Failed        int                       `json:"failed"`
	ProcessedList []ProcessedNotification   `json:"processed_list,omitempty"`
	FailedList    []FailedNotification      `json:"failed_list,omitempty"`
}

type ProcessedNotification struct {
	SequenceNumber int  `json:"sequence_number"`
	Removed        bool `json:"removed"`
}

type FailedNotification struct {
	SequenceNumber int    `json:"sequence_number"`
	Error          string `json:"error"`
}

type DiscoveryData struct {
	EventID string `json:"event_id"`
	Address string `json:"address"`
}

func NewHermesClient(config *HermesConfig) *HermesClient {
	return &HermesClient{config: config}
}

func (h *HermesClient) buildArgs() []string {
	var args []string
	if h.config.Driver != "" {
		args = append(args, "-driver", h.config.Driver)
	}
	if h.config.Device != "" {
		args = append(args, "-device", h.config.Device)
	}
	if h.config.Slot > 0 {
		args = append(args, "-slot", fmt.Sprintf("%d", h.config.Slot))
	}
	if h.config.Timeout > 0 {
		args = append(args, "-timeout", fmt.Sprintf("%d", h.config.Timeout))
	}
	return args
}

func (h *HermesClient) exec(cmdArgs ...string) (*HermesResponse, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	args := h.buildArgs()
	args = append(args, cmdArgs...)

	timeout := time.Duration(h.config.Timeout+10) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, h.config.Binary, args...)
	slog.Debug("hermes exec", "cmd", cmd.String())

	output, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("command timed out after %v", timeout)
		}
		if exitErr, ok := err.(*exec.ExitError); ok {
			var resp HermesResponse
			if json.Unmarshal(output, &resp) == nil {
				return &resp, nil
			}
			if json.Unmarshal(exitErr.Stderr, &resp) == nil {
				return &resp, nil
			}
			return nil, fmt.Errorf("command failed: %s", string(output))
		}
		return nil, fmt.Errorf("exec error: %w", err)
	}

	var resp HermesResponse
	if err := json.Unmarshal(output, &resp); err != nil {
		return nil, fmt.Errorf("invalid JSON response: %w", err)
	}
	return &resp, nil
}

func (h *HermesClient) ChipInfo() (*ChipInfoData, error) {
	resp, err := h.exec("chip-info")
	if err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf("%s", resp.Error)
	}
	var data ChipInfoData
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("parse chip-info: %w", err)
	}
	return &data, nil
}

func (h *HermesClient) EID() (string, error) {
	resp, err := h.exec("eid")
	if err != nil {
		return "", err
	}
	if !resp.Success {
		return "", fmt.Errorf("%s", resp.Error)
	}
	var data struct {
		EID string `json:"eid"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return "", fmt.Errorf("parse eid: %w", err)
	}
	return data.EID, nil
}

func (h *HermesClient) ListProfiles() ([]ProfileData, error) {
	resp, err := h.exec("list")
	if err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf("%s", resp.Error)
	}
	var profiles []ProfileData
	if err := json.Unmarshal(resp.Data, &profiles); err != nil {
		return nil, fmt.Errorf("parse profiles: %w", err)
	}
	return profiles, nil
}

func (h *HermesClient) EnableProfile(iccid string) error {
	resp, err := h.exec("enable", iccid)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("%s", resp.Error)
	}
	return nil
}

func (h *HermesClient) DisableProfile(iccid string) error {
	resp, err := h.exec("disable", iccid)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("%s", resp.Error)
	}
	return nil
}

func (h *HermesClient) DeleteProfile(iccid string) error {
	resp, err := h.exec("delete", iccid)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("%s", resp.Error)
	}
	return nil
}

func (h *HermesClient) SetNickname(iccid, nickname string) error {
	resp, err := h.exec("nickname", iccid, nickname)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("%s", resp.Error)
	}
	return nil
}

func (h *HermesClient) ListNotifications() ([]NotificationData, error) {
	resp, err := h.exec("notifications")
	if err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf("%s", resp.Error)
	}
	var notifications []NotificationData
	if err := json.Unmarshal(resp.Data, &notifications); err != nil {
		return nil, fmt.Errorf("parse notifications: %w", err)
	}
	return notifications, nil
}

func (h *HermesClient) HandleNotification(seqNumber string) error {
	resp, err := h.exec("notification-handle", seqNumber)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("%s", resp.Error)
	}
	return nil
}

func (h *HermesClient) RemoveNotification(seqNumber string) error {
	resp, err := h.exec("notification-remove", seqNumber)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("%s", resp.Error)
	}
	return nil
}

func (h *HermesClient) AutoNotification() (*AutoNotificationResult, error) {
	resp, err := h.exec("auto-notification")
	if err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf("%s", resp.Error)
	}
	var result AutoNotificationResult
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("parse auto-notification: %w", err)
	}
	return &result, nil
}

func (h *HermesClient) Download(activationCode, confirmationCode string) error {
	args := []string{"download", "--code", activationCode}
	if confirmationCode != "" {
		args = append(args, "--confirmation-code", confirmationCode)
	}
	resp, err := h.exec(args...)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("%s", resp.Error)
	}
	return nil
}

func (h *HermesClient) Discovery(server string) ([]DiscoveryData, error) {
	args := []string{"discovery"}
	if server != "" {
		args = append(args, "--server", server)
	}
	resp, err := h.exec(args...)
	if err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf("%s", resp.Error)
	}
	var discoveries []DiscoveryData
	if err := json.Unmarshal(resp.Data, &discoveries); err != nil {
		return nil, fmt.Errorf("parse discovery: %w", err)
	}
	return discoveries, nil
}

func (h *HermesClient) ConfiguredAddresses() (*ConfiguredAddresses, error) {
	resp, err := h.exec("configured-addresses")
	if err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf("%s", resp.Error)
	}
	var data ConfiguredAddresses
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("parse configured-addresses: %w", err)
	}
	return &data, nil
}

func (h *HermesClient) SetDefaultDP(address string) error {
	resp, err := h.exec("set-default-dp", address)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("%s", resp.Error)
	}
	return nil
}

func (h *HermesClient) Challenge() (string, error) {
	resp, err := h.exec("challenge")
	if err != nil {
		return "", err
	}
	if !resp.Success {
		return "", fmt.Errorf("%s", resp.Error)
	}
	var data struct {
		Challenge string `json:"challenge"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return "", fmt.Errorf("parse challenge: %w", err)
	}
	return data.Challenge, nil
}

func (h *HermesClient) MemoryReset() error {
	resp, err := h.exec("memory-reset")
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("%s", resp.Error)
	}
	return nil
}
