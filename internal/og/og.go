package og

import (
	"context"
	"strings"

	"Sixth_world_Suday/internal/config"
	"Sixth_world_Suday/internal/settings"
)

const (
	baseURLPlaceholder = "__BASE_URL__"
	imagePlaceholder   = "__OG_IMAGE__"
	namePlaceholder    = "__SITE_NAME__"
	descPlaceholder    = "__SITE_DESCRIPTION__"
)

type (
	Resolver struct {
		settingsSvc settings.Service
		baseHTML    string
		baseURL     string
	}
)

func NewResolver(settingsSvc settings.Service, baseHTML, baseURL string) *Resolver {
	return &Resolver{
		settingsSvc: settingsSvc,
		baseHTML:    strings.ReplaceAll(baseHTML, baseURLPlaceholder, baseURL),
		baseURL:     baseURL,
	}
}

func (r *Resolver) Resolve(ctx context.Context, _ string) string {
	siteName := r.settingsSvc.Get(ctx, config.SettingSiteName)
	siteDesc := r.settingsSvc.Get(ctx, config.SettingSiteDescription)
	image := r.settingsSvc.Get(ctx, config.SettingOGDefaultImage)

	html := r.baseHTML
	html = strings.ReplaceAll(html, namePlaceholder, escapeAttr(siteName))
	html = strings.ReplaceAll(html, descPlaceholder, escapeAttr(siteDesc))

	if image == "" {
		return stripImageTags(html)
	}

	return strings.ReplaceAll(html, imagePlaceholder, escapeAttr(r.publicImageURL(image)))
}

func (r *Resolver) publicImageURL(img string) string {
	if strings.HasPrefix(img, "http://") || strings.HasPrefix(img, "https://") {
		return img
	}

	const uploads = "/uploads/"
	if strings.HasPrefix(img, uploads) {
		return r.baseURL + "/og-image/" + strings.TrimPrefix(img, uploads)
	}

	return r.baseURL + "/" + strings.TrimPrefix(img, "/")
}

func stripImageTags(html string) string {
	html = stripMetaTag(html, "property", "og:image")
	html = stripMetaTag(html, "name", "twitter:image")
	return html
}

func escapeAttr(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

func stripMetaTag(html, attrName, attrValue string) string {
	prefix := `<meta ` + attrName + `="` + attrValue + `" content="`
	idx := strings.Index(html, prefix)
	if idx < 0 {
		return html
	}

	end := strings.Index(html[idx:], ">")
	if end < 0 {
		return html
	}

	return html[:idx] + html[idx+end+1:]
}
