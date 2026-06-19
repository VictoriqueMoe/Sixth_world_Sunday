package user

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClampDisplayName_PassThrough(t *testing.T) {
	// given
	in := "Cipher Augustus Aurora"

	// when
	out := ClampDisplayName(in)

	// then
	assert.Equal(t, in, out)
}

func TestClampDisplayName_TrimsLeadingTrailingWhitespace(t *testing.T) {
	// given
	in := "   Static   "

	// when
	out := ClampDisplayName(in)

	// then
	assert.Equal(t, "Static", out)
}

func TestClampDisplayName_CollapsesInternalWhitespace(t *testing.T) {
	// given
	in := "Raven\t\t\tWraith\n\nthe   decker"

	// when
	out := ClampDisplayName(in)

	// then
	assert.Equal(t, "Raven Wraith the decker", out)
}

func TestClampDisplayName_StripsHTMLTags(t *testing.T) {
	// given
	in := `<img src="x" alt="🔴">Nightjar<script>alert(1)</script>`

	// when
	out := ClampDisplayName(in)

	// then
	assert.Equal(t, "Nightjar", out)
}

func TestClampDisplayName_StripsBrokenImgWithEmojiAlt(t *testing.T) {
	// given
	in := `Cipher<img alt="🔴🟠🟡🟢🔵🟣🟤⚫⚪" src="">`

	// when
	out := ClampDisplayName(in)

	// then
	assert.Equal(t, "Cipher", out)
}

func TestClampDisplayName_CapsAt40Runes(t *testing.T) {
	// given
	in := strings.Repeat("a", 60)

	// when
	out := ClampDisplayName(in)

	// then
	assert.Equal(t, strings.Repeat("a", 40), out)
	assert.Equal(t, 40, len([]rune(out)))
}

func TestClampDisplayName_CapsByRunesNotBytes(t *testing.T) {
	// given: 60 multi-byte emoji (4 bytes each in UTF-8)
	in := strings.Repeat("🔴", 60)

	// when
	out := ClampDisplayName(in)

	// then: should keep 40 emoji, not 40 bytes
	assert.Equal(t, 40, len([]rune(out)))
	assert.Equal(t, strings.Repeat("🔴", 40), out)
}

func TestClampDisplayName_EmptyInput(t *testing.T) {
	// given
	in := ""

	// when
	out := ClampDisplayName(in)

	// then
	assert.Equal(t, "", out)
}

func TestClampDisplayName_OnlyWhitespace(t *testing.T) {
	// given
	in := "   \t\n  "

	// when
	out := ClampDisplayName(in)

	// then
	assert.Equal(t, "", out)
}

func TestClampDisplayName_OnlyHTML(t *testing.T) {
	// given
	in := "<b></b><img src=''><span></span>"

	// when
	out := ClampDisplayName(in)

	// then
	assert.Equal(t, "", out)
}

func TestClampDisplayName_StripsHTMLThenClampsRunes(t *testing.T) {
	// given: the HTML alt would push us over 40 if it weren't stripped first
	in := `<img alt="` + strings.Repeat("🔴", 30) + `">RealName`

	// when
	out := ClampDisplayName(in)

	// then
	assert.Equal(t, "RealName", out)
}

func TestClampDisplayName_HTMLEntitiesLeftEscaped(t *testing.T) {
	// given
	in := "Hatchet &amp; Nightjar"

	// when
	out := ClampDisplayName(in)

	// then
	assert.Equal(t, "Hatchet &amp; Nightjar", out)
}

func TestClampDisplayName_DropsScriptContent(t *testing.T) {
	// given: bluemonday's strict policy removes the script element AND its text.
	in := `Hatchet<script>alert("xss")</script>`

	// when
	out := ClampDisplayName(in)

	// then
	assert.Equal(t, "Hatchet", out)
}

func TestClampDisplayName_LongNameWithEmbeddedTags(t *testing.T) {
	// given
	in := `<b>` + strings.Repeat("Cipher ", 10) + `</b>`

	// when
	out := ClampDisplayName(in)

	// then
	assert.LessOrEqual(t, len([]rune(out)), 40)
	assert.False(t, strings.Contains(out, "<"))
}
