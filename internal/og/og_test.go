package og

import (
	"context"
	"strings"
	"testing"

	"Sixth_world_Suday/internal/config"
	"Sixth_world_Suday/internal/settings"

	"github.com/stretchr/testify/mock"
)

const sampleHTML = `<title>__SITE_NAME__</title>` +
	`<link rel="canonical" href="__BASE_URL__/">` +
	`<meta name="description" content="__SITE_DESCRIPTION__">` +
	`<meta property="og:title" content="__SITE_NAME__">` +
	`<meta property="og:description" content="__SITE_DESCRIPTION__">` +
	`<meta property="og:site_name" content="__SITE_NAME__">` +
	`<meta property="og:url" content="__BASE_URL__/">` +
	`<meta property="og:image" content="__OG_IMAGE__">` +
	`<meta name="twitter:image" content="__OG_IMAGE__">`

func TestResolver_InjectsSiteMetaAndStripsImageWhenUnset(t *testing.T) {
	ss := settings.NewMockService(t)
	ss.EXPECT().Get(mock.Anything, config.SettingSiteName).Return("Night City Net")
	ss.EXPECT().Get(mock.Anything, config.SettingSiteDescription).Return(`chrome & <shadows>`)
	ss.EXPECT().Get(mock.Anything, config.SettingOGDefaultImage).Return("")

	r := NewResolver(ss, sampleHTML, "https://sws.example")
	out := r.Resolve(context.Background(), "/channels/abc")

	if strings.Contains(out, "__BASE_URL__") || strings.Contains(out, "__SITE_NAME__") || strings.Contains(out, "__SITE_DESCRIPTION__") {
		t.Fatalf("placeholders not replaced: %s", out)
	}
	if !strings.Contains(out, "<title>Night City Net</title>") {
		t.Fatalf("title not injected: %s", out)
	}
	if !strings.Contains(out, `content="Night City Net"`) || !strings.Contains(out, `href="https://sws.example/"`) {
		t.Fatalf("site name / base url not injected: %s", out)
	}
	if !strings.Contains(out, "chrome &amp; &lt;shadows&gt;") {
		t.Fatalf("description not escaped/injected: %s", out)
	}
	if strings.Contains(out, "og:image") || strings.Contains(out, "twitter:image") {
		t.Fatalf("image tags should be stripped when no image set: %s", out)
	}
}

func TestResolver_InjectsPublicImageURL(t *testing.T) {
	ss := settings.NewMockService(t)
	ss.EXPECT().Get(mock.Anything, config.SettingSiteName).Return("X")
	ss.EXPECT().Get(mock.Anything, config.SettingSiteDescription).Return("Y")
	ss.EXPECT().Get(mock.Anything, config.SettingOGDefaultImage).Return("/uploads/branding/og_default_1.jpg")

	r := NewResolver(ss, sampleHTML, "https://sws.example")
	out := r.Resolve(context.Background(), "/")

	want := "https://sws.example/og-image/branding/og_default_1.jpg"
	if !strings.Contains(out, want) {
		t.Fatalf("expected public og-image URL %q in: %s", want, out)
	}
	if strings.Contains(out, "__OG_IMAGE__") {
		t.Fatalf("__OG_IMAGE__ placeholder not replaced: %s", out)
	}
	if strings.Contains(out, "/uploads/") {
		t.Fatalf("og:image must not point at the auth-gated /uploads path: %s", out)
	}
}
