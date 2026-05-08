// Copyright (c) 2025 Kilimcinin Kör Oğlu <k@keremgok.tr>
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"net/http"
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
	return "light"
}

func langFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(ctxLang).(string); ok {
		return v
	}
	return defaultLang
}

func preferencesMiddleware(i18n *I18n, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		theme := "light"
		if c, err := r.Cookie(cookieTheme); err == nil {
			if c.Value == "dark" || c.Value == "light" {
				theme = c.Value
			}
		}

		lang := defaultLang
		if c, err := r.Cookie(cookieLang); err == nil {
			if i18n.IsSupported(c.Value) {
				lang = c.Value
			}
		}

		ctx := context.WithValue(r.Context(), ctxTheme, theme)
		ctx = context.WithValue(ctx, ctxLang, lang)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
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
