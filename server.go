// Copyright (c) 2025 Kilimcinin Kör Oğlu <k@keremgok.tr>
// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"embed"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
)

//go:embed static
var staticFS embed.FS

//go:embed templates
var templatesFS embed.FS

type Server struct {
	mux      *http.ServeMux
	i18n     *I18n
	hermes   *HermesClient
	tmpl     *template.Template
	handler  http.Handler
}

type HermesConfig struct {
	Binary  string
	Driver  string
	Device  string
	Slot    int
	Timeout int
}

type PageData struct {
	Theme       string
	Lang        string
	Languages   []LangInfo
	ActivePage  string
	T           func(string) template.HTML
	Version     string
	PageContent template.HTML
	Data        any
}

func NewServer(i18n *I18n, hermes *HermesClient) *Server {
	s := &Server{
		mux:    http.NewServeMux(),
		i18n:   i18n,
		hermes: hermes,
	}

	s.parseTemplates()
	s.registerRoutes()
	s.handler = preferencesMiddleware(i18n, s.mux)
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.handler.ServeHTTP(w, r)
}

func (s *Server) parseTemplates() {
	funcMap := template.FuncMap{
		"t":  func(key string) template.HTML { return "" },
		"eq": func(a, b string) bool { return a == b },
	}

	tmpl, err := template.New("").Funcs(funcMap).ParseFS(templatesFS,
		"templates/layout.html",
		"templates/components/*.html",
		"templates/pages/*.html",
	)
	if err != nil {
		slog.Error("failed to parse templates", "error", err)
		panic(err)
	}
	s.tmpl = tmpl
}

func (s *Server) registerRoutes() {
	staticSub, _ := fs.Sub(staticFS, "static")
	s.mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticSub))))

	s.mux.HandleFunc("GET /", s.handlePage("info"))
	s.mux.HandleFunc("GET /profiles", s.handlePage("profiles"))
	s.mux.HandleFunc("GET /download", s.handlePage("download"))
	s.mux.HandleFunc("GET /notifications", s.handlePage("notifications"))
	s.mux.HandleFunc("GET /settings", s.handlePage("settings"))

	s.mux.HandleFunc("POST /api/theme", s.handleSetTheme)
	s.mux.HandleFunc("POST /api/lang", s.handleSetLang)

	s.mux.HandleFunc("GET /api/info", s.handleAPIInfo)
	s.mux.HandleFunc("GET /api/profiles", s.handleAPIProfiles)
	s.mux.HandleFunc("POST /api/profiles/{iccid}/enable", s.handleAPIProfileEnable)
	s.mux.HandleFunc("POST /api/profiles/{iccid}/disable", s.handleAPIProfileDisable)
	s.mux.HandleFunc("POST /api/profiles/{iccid}/delete", s.handleAPIProfileDelete)
	s.mux.HandleFunc("POST /api/profiles/{iccid}/nickname", s.handleAPIProfileNickname)

	s.mux.HandleFunc("POST /api/download", s.handleAPIDownload)
	s.mux.HandleFunc("POST /api/discovery", s.handleAPIDiscovery)

	s.mux.HandleFunc("GET /api/notifications", s.handleAPINotifications)
	s.mux.HandleFunc("POST /api/notifications/{seq}/handle", s.handleAPINotificationHandle)
	s.mux.HandleFunc("POST /api/notifications/{seq}/remove", s.handleAPINotificationRemove)
	s.mux.HandleFunc("POST /api/notifications/process-all", s.handleAPINotificationProcessAll)

	s.mux.HandleFunc("GET /api/settings", s.handleAPISettingsGet)
	s.mux.HandleFunc("POST /api/settings", s.handleAPISettingsSave)

	s.mux.HandleFunc("POST /api/info/set-default-dp", s.handleAPISetDefaultDP)
	s.mux.HandleFunc("GET /api/info/challenge", s.handleAPIChallenge)
	s.mux.HandleFunc("POST /api/info/memory-reset", s.handleAPIMemoryReset)

	s.mux.HandleFunc("POST /api/reboot", s.handleAPIReboot)
	s.mux.HandleFunc("GET /api/connectivity", s.handleAPIConnectivity)
	s.mux.HandleFunc("GET /api/storage-check", s.handleAPIStorageCheck)
}

func (s *Server) handlePage(page string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if page == "info" && r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		lang := langFromContext(r.Context())
		theme := themeFromContext(r.Context())
		isHTMX := r.Header.Get("HX-Request") == "true"

		data := &PageData{
			Theme:      theme,
			Lang:       lang,
			Languages:  s.i18n.Languages(),
			ActivePage: page,
			T:          s.i18n.TranslateFunc(lang),
			Version:    version,
		}

		tmpl := s.tmpl.Funcs(template.FuncMap{
			"t": s.i18n.TranslateFunc(lang),
		})

		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		if isHTMX {
			if err := tmpl.ExecuteTemplate(w, page, data); err != nil {
				slog.Error("template render error", "page", page, "error", err)
				http.Error(w, "render error", http.StatusInternalServerError)
			}
			return
		}

		var buf bytes.Buffer
		if err := tmpl.ExecuteTemplate(&buf, page, data); err != nil {
			slog.Error("template render error", "page", page, "error", err)
			http.Error(w, "render error", http.StatusInternalServerError)
			return
		}
		data.PageContent = template.HTML(buf.String())

		if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
			slog.Error("layout render error", "page", page, "error", err)
			http.Error(w, "render error", http.StatusInternalServerError)
		}
	}
}

func (s *Server) handleSetTheme(w http.ResponseWriter, r *http.Request) {
	theme := r.FormValue("theme")
	if theme != "dark" && theme != "light" {
		theme = "light"
	}
	setPreferenceCookie(w, cookieTheme, theme)
	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleSetLang(w http.ResponseWriter, r *http.Request) {
	lang := r.FormValue("lang")
	if !s.i18n.IsSupported(lang) {
		lang = defaultLang
	}
	setPreferenceCookie(w, cookieLang, lang)
	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}
