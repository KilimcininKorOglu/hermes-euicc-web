// Copyright (c) 2025 Kilimcinin Kör Oğlu <k@keremgok.tr>
// SPDX-License-Identifier: MIT

package main

import (
	"html/template"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

func (s *Server) handleAPIInfo(w http.ResponseWriter, r *http.Request) {
	lang := langFromContext(r.Context())

	chipInfo, err := s.hermes.ChipInfo()
	if err != nil {
		s.renderFragment(w, r, "info_error", &PageData{
			T:    s.i18n.TranslateFunc(lang),
			Data: err.Error(),
		})
		return
	}

	s.renderFragment(w, r, "info_content", &PageData{
		T:    s.i18n.TranslateFunc(lang),
		Data: chipInfo,
	})
}

func (s *Server) handleAPIProfiles(w http.ResponseWriter, r *http.Request) {
	lang := langFromContext(r.Context())

	profiles, err := s.hermes.ListProfiles()
	if err != nil {
		s.renderFragment(w, r, "profiles_error", &PageData{
			T:    s.i18n.TranslateFunc(lang),
			Data: err.Error(),
		})
		return
	}

	s.renderFragment(w, r, "profiles_table", &PageData{
		T: s.i18n.TranslateFunc(lang),
		Data: &profilesData{
			Profiles: profiles,
		},
	})
}

type profilesData struct {
	Profiles     []ProfileData
	RebootNeeded bool
}

func (s *Server) handleAPIProfilesWithReboot(w http.ResponseWriter, r *http.Request) {
	lang := langFromContext(r.Context())

	profiles, err := s.hermes.ListProfiles()
	if err != nil {
		s.renderFragment(w, r, "profiles_error", &PageData{
			T:    s.i18n.TranslateFunc(lang),
			Data: err.Error(),
		})
		return
	}

	s.renderFragment(w, r, "profiles_table", &PageData{
		T: s.i18n.TranslateFunc(lang),
		Data: &profilesData{
			Profiles:     profiles,
			RebootNeeded: true,
		},
	})
}

func (s *Server) handleAPIProfileEnable(w http.ResponseWriter, r *http.Request) {
	iccid := r.PathValue("iccid")
	if iccid == "" {
		http.Error(w, "missing iccid", http.StatusBadRequest)
		return
	}

	if err := s.hermes.EnableProfile(iccid); err != nil {
		slog.Error("enable profile failed", "iccid", iccid, "error", err)
		s.renderAlertFragment(w, r, "danger", err.Error())
		return
	}

	s.handleAPIProfilesWithReboot(w, r)
}

func (s *Server) handleAPIProfileDisable(w http.ResponseWriter, r *http.Request) {
	iccid := r.PathValue("iccid")
	if iccid == "" {
		http.Error(w, "missing iccid", http.StatusBadRequest)
		return
	}

	if err := s.hermes.DisableProfile(iccid); err != nil {
		slog.Error("disable profile failed", "iccid", iccid, "error", err)
		s.renderAlertFragment(w, r, "danger", err.Error())
		return
	}

	s.handleAPIProfilesWithReboot(w, r)
}

func (s *Server) handleAPIProfileDelete(w http.ResponseWriter, r *http.Request) {
	iccid := r.PathValue("iccid")
	if iccid == "" {
		http.Error(w, "missing iccid", http.StatusBadRequest)
		return
	}

	if err := s.hermes.DeleteProfile(iccid); err != nil {
		slog.Error("delete profile failed", "iccid", iccid, "error", err)
		s.renderAlertFragment(w, r, "danger", err.Error())
		return
	}

	s.handleAPIProfiles(w, r)
}

func (s *Server) handleAPIProfileNickname(w http.ResponseWriter, r *http.Request) {
	iccid := r.PathValue("iccid")
	nickname := r.FormValue("nickname")
	if iccid == "" {
		http.Error(w, "missing iccid", http.StatusBadRequest)
		return
	}

	if err := s.hermes.SetNickname(iccid, nickname); err != nil {
		slog.Error("set nickname failed", "iccid", iccid, "error", err)
		s.renderAlertFragment(w, r, "danger", err.Error())
		return
	}

	s.handleAPIProfiles(w, r)
}

func (s *Server) handleAPIDownload(w http.ResponseWriter, r *http.Request) {
	lang := langFromContext(r.Context())
	activationCode := r.FormValue("activation_code")
	confirmationCode := r.FormValue("confirmation_code")

	if activationCode == "" {
		s.renderAlertFragment(w, r, "danger", s.i18n.Translate(lang, "download.error_no_code"))
		return
	}

	if err := s.hermes.Download(activationCode, confirmationCode); err != nil {
		slog.Error("download failed", "error", err)
		s.renderFragment(w, r, "download_result", &PageData{
			T: s.i18n.TranslateFunc(lang),
			Data: map[string]string{
				"status":  "error",
				"message": err.Error(),
			},
		})
		return
	}

	s.renderFragment(w, r, "download_result", &PageData{
		T: s.i18n.TranslateFunc(lang),
		Data: map[string]string{
			"status":  "success",
			"message": s.i18n.Translate(lang, "download.success"),
		},
	})
}

func (s *Server) handleAPIDiscovery(w http.ResponseWriter, r *http.Request) {
	lang := langFromContext(r.Context())
	server := r.FormValue("server")

	discoveries, err := s.hermes.Discovery(server)
	if err != nil {
		slog.Error("discovery failed", "error", err)
		s.renderFragment(w, r, "discovery_result", &PageData{
			T: s.i18n.TranslateFunc(lang),
			Data: map[string]any{
				"error": err.Error(),
			},
		})
		return
	}

	s.renderFragment(w, r, "discovery_result", &PageData{
		T: s.i18n.TranslateFunc(lang),
		Data: map[string]any{
			"discoveries": discoveries,
		},
	})
}

func (s *Server) handleAPINotifications(w http.ResponseWriter, r *http.Request) {
	lang := langFromContext(r.Context())

	notifications, err := s.hermes.ListNotifications()
	if err != nil {
		s.renderFragment(w, r, "notifications_error", &PageData{
			T:    s.i18n.TranslateFunc(lang),
			Data: err.Error(),
		})
		return
	}

	s.renderFragment(w, r, "notifications_table", &PageData{
		T:    s.i18n.TranslateFunc(lang),
		Data: notifications,
	})
}

func (s *Server) handleAPINotificationHandle(w http.ResponseWriter, r *http.Request) {
	seq := r.PathValue("seq")
	if seq == "" {
		http.Error(w, "missing seq", http.StatusBadRequest)
		return
	}

	if err := s.hermes.HandleNotification(seq); err != nil {
		slog.Error("handle notification failed", "seq", seq, "error", err)
		s.renderAlertFragment(w, r, "danger", err.Error())
		return
	}

	s.handleAPINotifications(w, r)
}

func (s *Server) handleAPINotificationRemove(w http.ResponseWriter, r *http.Request) {
	seq := r.PathValue("seq")
	if seq == "" {
		http.Error(w, "missing seq", http.StatusBadRequest)
		return
	}

	if err := s.hermes.RemoveNotification(seq); err != nil {
		slog.Error("remove notification failed", "seq", seq, "error", err)
		s.renderAlertFragment(w, r, "danger", err.Error())
		return
	}

	s.handleAPINotifications(w, r)
}

func (s *Server) handleAPINotificationProcessAll(w http.ResponseWriter, r *http.Request) {
	lang := langFromContext(r.Context())

	result, err := s.hermes.AutoNotification()
	if err != nil {
		slog.Error("auto-notification failed", "error", err)
		s.renderAlertFragment(w, r, "danger", err.Error())
		return
	}

	s.renderFragment(w, r, "notifications_process_result", &PageData{
		T:    s.i18n.TranslateFunc(lang),
		Data: result,
	})
}

func (s *Server) handleAPISettingsGet(w http.ResponseWriter, r *http.Request) {
	lang := langFromContext(r.Context())

	config, err := loadConfig()
	if err != nil {
		slog.Error("load config failed", "error", err)
		config = &AppConfig{Driver: "auto", Slot: 1, Timeout: 30}
	}

	s.renderFragment(w, r, "settings_form", &PageData{
		T:    s.i18n.TranslateFunc(lang),
		Data: config,
	})
}

func (s *Server) handleAPISettingsSave(w http.ResponseWriter, r *http.Request) {
	lang := langFromContext(r.Context())

	slot, _ := strconv.Atoi(r.FormValue("slot"))
	if slot < 1 {
		slot = 1
	}
	timeout, _ := strconv.Atoi(r.FormValue("timeout"))
	if timeout < 1 {
		timeout = 30
	}

	driver := r.FormValue("driver")
	validDrivers := map[string]bool{"auto": true, "qmi": true, "mbim": true, "at": true, "ccid": true}
	if !validDrivers[driver] {
		driver = "auto"
	}

	rebootQMISlot, _ := strconv.Atoi(r.FormValue("reboot_qmi_slot"))
	if rebootQMISlot < 1 {
		rebootQMISlot = 1
	}

	rebootMethod := r.FormValue("reboot_method")
	validReboot := map[string]bool{"none": true, "at": true, "qmi": true, "mbim": true, "custom": true}
	if !validReboot[rebootMethod] {
		rebootMethod = "none"
	}

	config := &AppConfig{
		Driver:              driver,
		Device:              r.FormValue("device"),
		Slot:                slot,
		Timeout:             timeout,
		EnableOutputLogs:    r.FormValue("enable_output_logs") == "on",
		AutoNotification:    r.FormValue("auto_notification") == "on",
		RebootMethod:        rebootMethod,
		RebootATCommand:     r.FormValue("reboot_at_command"),
		RebootATDevice:      r.FormValue("reboot_at_device"),
		RebootQMIDevice:     r.FormValue("reboot_qmi_device"),
		RebootQMISlot:       rebootQMISlot,
		RebootMBIMDevice:    r.FormValue("reboot_mbim_device"),
		RebootCustomCommand: r.FormValue("reboot_custom_command"),
	}

	if err := saveConfig(config); err != nil {
		slog.Error("save config failed", "error", err)
		s.renderAlertFragment(w, r, "danger", err.Error())
		return
	}

	s.renderFragment(w, r, "settings_saved", &PageData{
		T:    s.i18n.TranslateFunc(lang),
		Data: config,
	})
}

func (s *Server) handleAPISetDefaultDP(w http.ResponseWriter, r *http.Request) {
	lang := langFromContext(r.Context())
	address := r.FormValue("address")
	if address == "" {
		s.renderAlertFragment(w, r, "danger", s.i18n.Translate(lang, "info.error_no_address"))
		return
	}

	if err := s.hermes.SetDefaultDP(address); err != nil {
		slog.Error("set default dp failed", "error", err)
		s.renderAlertFragment(w, r, "danger", err.Error())
		return
	}

	s.renderAlertFragment(w, r, "success", s.i18n.Translate(lang, "info.default_dp_set"))
}

func (s *Server) handleAPIChallenge(w http.ResponseWriter, r *http.Request) {
	lang := langFromContext(r.Context())

	challenge, err := s.hermes.Challenge()
	if err != nil {
		slog.Error("challenge failed", "error", err)
		s.renderAlertFragment(w, r, "danger", err.Error())
		return
	}

	s.renderFragment(w, r, "challenge_result", &PageData{
		T:    s.i18n.TranslateFunc(lang),
		Data: challenge,
	})
}

func (s *Server) handleAPIMemoryReset(w http.ResponseWriter, r *http.Request) {
	lang := langFromContext(r.Context())

	if err := s.hermes.MemoryReset(); err != nil {
		slog.Error("memory reset failed", "error", err)
		s.renderAlertFragment(w, r, "danger", err.Error())
		return
	}

	s.renderAlertFragment(w, r, "success", s.i18n.Translate(lang, "info.memory_reset_success"))
}

func (s *Server) handleAPIReboot(w http.ResponseWriter, r *http.Request) {
	lang := langFromContext(r.Context())

	config, err := loadConfig()
	if err != nil {
		s.renderAlertFragment(w, r, "danger", err.Error())
		return
	}

	if err := execReboot(config); err != nil {
		slog.Error("reboot failed", "error", err)
		s.renderAlertFragment(w, r, "danger", err.Error())
		return
	}

	s.renderAlertFragment(w, r, "success", s.i18n.Translate(lang, "profiles.reboot_success"))
}

func (s *Server) handleAPIConnectivity(w http.ResponseWriter, r *http.Request) {
	lang := langFromContext(r.Context())

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get("http://connectivitycheck.gstatic.com/generate_204")
	online := err == nil && resp != nil && resp.StatusCode == 204
	if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}

	s.renderFragment(w, r, "connectivity_banner", &PageData{
		T:    s.i18n.TranslateFunc(lang),
		Data: online,
	})
}

func (s *Server) handleAPIStorageCheck(w http.ResponseWriter, r *http.Request) {
	lang := langFromContext(r.Context())

	chipInfo, err := s.hermes.ChipInfo()
	if err != nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	var freeNVM int
	if chipInfo.EUICCInfo2 != nil && chipInfo.EUICCInfo2.ExtCardResource != nil {
		freeNVM = chipInfo.EUICCInfo2.ExtCardResource.FreeNonVolatileMemory
	}

	if freeNVM > 0 && freeNVM < 32768 {
		s.renderFragment(w, r, "storage_warning", &PageData{
			T:    s.i18n.TranslateFunc(lang),
			Data: freeNVM,
		})
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) renderFragment(w http.ResponseWriter, r *http.Request, name string, data *PageData) {
	lang := langFromContext(r.Context())
	if data.T == nil {
		data.T = s.i18n.TranslateFunc(lang)
	}

	tmpl := s.tmpl.Funcs(template.FuncMap{
		"t": s.i18n.TranslateFunc(lang),
	})

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, name, data); err != nil {
		slog.Error("fragment render error", "name", name, "error", err)
		http.Error(w, "render error", http.StatusInternalServerError)
	}
}

func (s *Server) renderAlertFragment(w http.ResponseWriter, r *http.Request, alertType, message string) {
	lang := langFromContext(r.Context())
	tmpl := s.tmpl.Funcs(template.FuncMap{
		"t": s.i18n.TranslateFunc(lang),
	})

	data := struct {
		AlertType string
		Message   string
	}{alertType, message}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "alert", &data); err != nil {
		slog.Error("alert render error", "error", err)
		http.Error(w, message, http.StatusInternalServerError)
	}
}
