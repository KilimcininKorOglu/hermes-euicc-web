// Copyright (c) 2025 Kilimcinin Kör Oğlu <k@keremgok.tr>
// SPDX-License-Identifier: MIT

package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"sort"
)

//go:embed locales/*.json
var localesFS embed.FS

var supportedLanguages = []LangInfo{
	{Code: "en", Name: "English"},
	{Code: "tr", Name: "Türkçe (Turkey)"},
	{Code: "de", Name: "Deutsch (Germany)"},
	{Code: "fr", Name: "Français (France)"},
	{Code: "fi", Name: "Suomi (Finland)"},
	{Code: "da", Name: "Dansk (Denmark)"},
	{Code: "pt", Name: "Português (Portugal)"},
	{Code: "es", Name: "Español (Spain)"},
	{Code: "it", Name: "Italiano (Italy)"},
	{Code: "nl", Name: "Nederlands (Netherlands)"},
	{Code: "pl", Name: "Polski (Poland)"},
	{Code: "ru", Name: "Русский (Russia)"},
	{Code: "zh", Name: "中文 (China)"},
}

const defaultLang = "en"

type LangInfo struct {
	Code string
	Name string
}

type I18n struct {
	translations map[string]map[string]string
	languages    []LangInfo
}

func NewI18n() (*I18n, error) {
	i := &I18n{
		translations: make(map[string]map[string]string),
		languages:    supportedLanguages,
	}

	for _, lang := range supportedLanguages {
		data, err := localesFS.ReadFile(fmt.Sprintf("locales/%s.json", lang.Code))
		if err != nil {
			return nil, fmt.Errorf("load locale %s: %w", lang.Code, err)
		}
		var msgs map[string]string
		if err := json.Unmarshal(data, &msgs); err != nil {
			return nil, fmt.Errorf("parse locale %s: %w", lang.Code, err)
		}
		i.translations[lang.Code] = msgs
		slog.Debug("loaded locale", "lang", lang.Code, "keys", len(msgs))
	}

	return i, nil
}

func (i *I18n) Translate(lang, key string) string {
	if msgs, ok := i.translations[lang]; ok {
		if val, ok := msgs[key]; ok {
			return val
		}
	}
	if lang != defaultLang {
		if msgs, ok := i.translations[defaultLang]; ok {
			if val, ok := msgs[key]; ok {
				return val
			}
		}
	}
	return key
}

func (i *I18n) IsSupported(lang string) bool {
	for _, l := range i.languages {
		if l.Code == lang {
			return true
		}
	}
	return false
}

func (i *I18n) Languages() []LangInfo {
	return i.languages
}

func (i *I18n) TranslateFunc(lang string) func(string) template.HTML {
	return func(key string) template.HTML {
		return template.HTML(template.HTMLEscapeString(i.Translate(lang, key)))
	}
}

func (i *I18n) AllKeys() []string {
	keySet := make(map[string]struct{})
	for _, msgs := range i.translations {
		for k := range msgs {
			keySet[k] = struct{}{}
		}
	}
	keys := make([]string, 0, len(keySet))
	for k := range keySet {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
