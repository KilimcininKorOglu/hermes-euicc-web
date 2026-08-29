// Copyright (c) 2025 Kilimcinin Kör Oğlu <k@keremgok.tr>
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const (
	ctxTheme contextKey = "theme"
	ctxLang  contextKey = "lang"
)

const (
	cookieTheme  = "theme"
	cookieLang   = "lang"
	cookieMaxAge = 365 * 24 * 60 * 60
)

func themeFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(ctxTheme).(string); ok {
		return v
	}
	return "system"
}

// isValidTheme reports whether v is an accepted theme preference.
func isValidTheme(v string) bool {
	return v == "light" || v == "dark" || v == "system"
}

func langFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(ctxLang).(string); ok {
		return v
	}
	return defaultLang
}

func preferencesMiddleware(i18n *I18n, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		theme := "system"
		if c, err := r.Cookie(cookieTheme); err == nil {
			if isValidTheme(c.Value) {
				theme = c.Value
			}
		}

		lang := defaultLang
		if c, err := r.Cookie(cookieLang); err == nil {
			if i18n.IsSupported(c.Value) {
				lang = c.Value
			}
		} else {
			lang = detectBrowserLang(r, i18n)
		}

		// API responses carry live eUICC/state data (some non-deterministic,
		// e.g. /api/info/challenge); never cache them.
		if strings.HasPrefix(r.URL.Path, "/api/") {
			w.Header().Set("Cache-Control", "no-store")
		}

		ctx := context.WithValue(r.Context(), ctxTheme, theme)
		ctx = context.WithValue(ctx, ctxLang, lang)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func detectBrowserLang(r *http.Request, i18n *I18n) string {
	accept := r.Header.Get("Accept-Language")
	if accept == "" {
		return defaultLang
	}
	for part := range strings.SplitSeq(accept, ",") {
		tag := strings.TrimSpace(strings.SplitN(part, ";", 2)[0])
		if i18n.IsSupported(tag) {
			return tag
		}
		if base, _, ok := strings.Cut(tag, "-"); ok && i18n.IsSupported(base) {
			return base
		}
	}
	return defaultLang
}

func setPreferenceCookie(w http.ResponseWriter, name, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   cookieMaxAge,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: false,
	})
}
