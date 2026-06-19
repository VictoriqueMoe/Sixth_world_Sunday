# Sixth World Sunday — "Neon Sprawl" Design Brief

The single source of truth for the frontend rebuild. Every screen must follow this. The reference mock is `design/sixth-world-sunday-mockup.html`. This is a **cyberpunk Shadowrun secure-node** aesthetic: dark, terminal/HUD, neon magenta + cyan, CRT texture. NOT generic, NOT City-of-Books.

## Tokens (use these CSS variables — never hardcode hex)

Surfaces: `--bg` #07070c · `--bg-2` #0b0b13 · `--panel` #0f1019 · `--panel-2` #14151f · `--panel-3` #191a26 · `--line` #20212f · `--line-bright` #2e3045
Text: `--text` #e9e9f2 · `--dim` #80819a · `--dimmer` #4d4e62
Accents: `--magenta` #ff2d78 (primary) · `--cyan` #00e5ff (secondary) · `--amber` #ffb000 · `--green` #38ffa3 · `--violet` #9d6bff · `--teal` #2ad6c4 · `--pink` #ff7ac6
RGB (for rgba glows): `--magenta-rgb`, `--cyan-rgb`, `--line-rgb`
Type: `--display` (Chakra Petch) for UI/body · `--mono` (Share Tech Mono) for labels/metadata/codes

The global body already paints the bg gradient-glow + cyan grid + CRT scanlines + vignette. Screens sit on transparent/`--panel` surfaces over it.

## Type rules
- Body/UI text: `--display`, ~14–15px, `letter-spacing: .2px`.
- **Labels, metadata, timestamps, counts, status, codes, channel glyphs**: `--mono`, small (10–12px), often `text-transform: uppercase` + `letter-spacing: 1–3px`, color `--dim`/`--dimmer`.
- Section headers: mono, uppercase, `--dimmer`, prefixed with `// ` (e.g. `// text channels`).
- Page/dialog titles (h2): uppercase, 600 weight, letter-spacing 2px, prefixed with a magenta `▸ ` glyph.
- Big brand/wordmark: Chakra Petch 700, uppercase; the "DAY" in SUNDAY is magenta with glow.

## Surfaces & borders
- Cards/panels: `background: var(--panel)`, `border: 1px solid var(--line)` (use `--line-bright` for emphasis). Mostly **square corners or 2px radius** — this is a HUD, avoid big rounded corners.
- Inputs/fields: `background: #07070d`, `border: 1px solid var(--line-bright)`, mono text; on focus → `border-color: var(--magenta)` + `box-shadow: 0 0 0 1px var(--magenta), 0 0 18px -4px var(--magenta)`.

## Buttons
- Primary ("jack in" style): `border: 1px solid var(--magenta)`, `background: linear-gradient(180deg, rgba(var(--magenta-rgb),.22), rgba(var(--magenta-rgb),.06))`, white text, uppercase, letter-spacing 2–3px; hover → `box-shadow: 0 0 28px -4px var(--magenta)`; active → translateY(1px). Optional diagonal sweep highlight.
- Secondary/ghost: transparent or `--panel-2` bg, `1px solid var(--line-bright)`, `--dim` text; hover → `--text` + `border-color: var(--cyan)`.
- Danger: hover → magenta border/text.

## Signature components
- **Channel row**: `# ` (text, mono `--dimmer`) or `◊ ` (voice) or `▤/▦/♪` (archive) glyph + name; `--dim` text; `border-left: 2px solid transparent`; hover → `--panel-2` + `--text`; active → `background: linear-gradient(90deg, rgba(var(--magenta-rgb),.14), transparent)` + white text + `border-left-color: var(--magenta)` + magenta glyph. Live badge: mono, magenta border, pulsing dot.
- **Role pill** (Shadowrun roles): mono 9.5px uppercase, `padding: 1px 5px`, `border: 1px solid` in the role color, text in role color. Role colors: Decker=magenta, Street Samurai=cyan, Fixer=amber, Mage=violet, Rigger=green, Technomancer=teal, Face=pink.
- **Avatar**: square, `border-radius: 3–4px`, 1px solid border tinted to role color, mono initial. Presence dot bottom-right (green on / amber idle / dimmer off).
- **Status badge**: mono uppercase, a small dot (green = ok with glow, cyan, amber = warn) + label.
- **Message**: square avatar + body; head = name (role color) + role pill + mono timestamp; mentions = `--cyan` on `rgba(cyan,.10)`; inline code = mono, amber, dark bg.
- **Telemetry/HUD lists**: mono, key left in `--dim`, value right colored (`--green` ok, `--amber` warn, `--cyan` info), dashed `--line-bright` divider.

## Motion (subtle, purposeful)
- `pulse` (status dots), `blink` (cursor), `sweep` (button highlight), `flicker` (wordmark), `rise` (card entrance). Keep tasteful; one orchestrated entrance per view max.

## Copy / voice (Shadowrun)
"Jack In" (login), "Handle / SIN" + "Passkey" (credentials), "transmit to #channel" (composer placeholder), "the Archive", "jacked in" (online), "node online", "runners" (users), "the grid" (search). Channels use `#name` (text) and `◊ Name` (voice). Replace any Umineko/CoB copy.

## Layout (the chat shell, from the mock)
Topbar (~46px): brand `SIXTH WORLD SUNDAY` (DAY magenta), node-status badges, spacer, search, me. Left sidebar (~268px): server head + scrollable channel groups (text / voice / archive) + voice dock pinned bottom + me-card. Center: chat head (title + topic) + messages (scroll) + composer. Right members rail (~232px): grouped "jacked in" / "offline" with avatars + role pills. Collapse rails on narrow viewports.

## Hard rules
- No City-of-Books / Umineko anything (no butterflies, truth classes, gold/witch theming, theory/mystery/etc.).
- Use tokens, not hex. Use `--mono` for all labels/meta. Square-ish HUD surfaces, neon accents, restrained glow.
- Keep all existing data wiring (hooks/queries/mutations, handlers, props) — rewrite presentation only.
