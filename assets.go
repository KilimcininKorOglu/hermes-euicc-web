// Copyright (c) 2025 Kilimcinin Kör Oğlu <k@keremgok.tr>
// SPDX-License-Identifier: MIT

package main

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"path"
	"strings"
)

// staticAsset holds a preloaded static file with its precomputed cache metadata.
type staticAsset struct {
	data         []byte
	etag         string // empty for immutable assets, where ETag is redundant
	cacheControl string
	version      string // short hash used for ?v= cache-busting in templates
}

// buildStaticAssets reads every file under fsys once at startup and computes a
// SHA-256 content hash, so per-request hashing is never needed. The map key is
// the path relative to fsys (e.g. "js/app.js").
func buildStaticAssets(fsys fs.FS) (map[string]staticAsset, error) {
	assets := make(map[string]staticAsset)
	err := fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		data, err := fs.ReadFile(fsys, p)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(data)
		hexSum := fmt.Sprintf("%x", sum)
		cacheControl, useETag := cacheControlFor(p)
		asset := staticAsset{
			data:         data,
			cacheControl: cacheControl,
			version:      hexSum[:12],
		}
		if useETag {
			asset.etag = `"` + hexSum + `"`
		}
		assets[p] = asset
		return nil
	})
	if err != nil {
		return nil, err
	}
	return assets, nil
}

// cacheControlFor picks a Cache-Control policy from the file name and reports
// whether an ETag should also be sent. Versioned library bundles are immutable
// and rely on ?v= busting, so they skip the ETag; own assets revalidate.
func cacheControlFor(name string) (string, bool) {
	base := path.Base(name)
	if strings.HasSuffix(base, ".min.js") || strings.HasSuffix(base, ".min.css") {
		return "public, max-age=31536000, immutable", false
	}
	switch path.Ext(base) {
	case ".svg", ".png", ".jpg", ".jpeg", ".gif", ".ico", ".webp", ".woff", ".woff2":
		return "public, max-age=86400", true
	default:
		return "public, max-age=3600", true
	}
}

// matchETag reports whether ifNoneMatch (a possibly comma-separated list, or "*")
// matches etag using RFC 7232 weak comparison.
func matchETag(ifNoneMatch, etag string) bool {
	if ifNoneMatch == "" || etag == "" {
		return false
	}
	if strings.TrimSpace(ifNoneMatch) == "*" {
		return true
	}
	candidate := strings.TrimPrefix(etag, "W/")
	for tag := range strings.SplitSeq(ifNoneMatch, ",") {
		tag = strings.TrimSpace(tag)
		tag = strings.TrimPrefix(tag, "W/")
		if tag == candidate {
			return true
		}
	}
	return false
}
