Task Result

Task ID: bg_ee247d70
Description: Audit design tokens and styling
Duration: 4m 27s
Session ID: ses_13cd48414ffeb61n50TKfpW1qx

---

<analysis>
**Literal Request**: Thorough scan of design-system definition vs actual code usage across 8 specific areas.
**Actual Need**: Precise gap analysis data with exact citations to build redesign brief.
**Success Looks Like**: Enumerated markdown report with absolute paths, line numbers, concrete values, no summarization.
</analysis>

Got design docs + app.css. Now do parallel audits for code usage.

Got enough data. Let me grab the rounded-[Npx] sample and a few more confirmations.

Have enough data. Now compile final report.

# JEDUG Design System — Gap Analysis Audit

## 1. Design Docs Inventory (`/Users/ryanprayoga/Project/Pribadi/jedug_v2/design-docs/`)

### `/Users/ryanprayoga/Project/Pribadi/jedug_v2/design-docs/design-system.md` (225 lines)

Source-of-truth for tokens. Defines:

- **Brand identity** (lines 16-27): runtime = Tailwind v4 + tokens in `frontend/src/app.css`; icons = Iconify Solar `line-duotone`; dark mode via `data-theme` + `.dark` class; persist key `jedug-theme`; modes `light|dark|system`.
- **Brand palette** (lines 31-37): `JEDUG Red #E5484D`, `Red Dep #C5363A`, `Red Soft #FF4F3`.
- **Severity** (lines 39-45): `1=#F6C453`, `2=#F97316`, `3+=#DC2626`.
- **Heatmap** (lines 47-54): 4 rgba steps yellow→dep red.
- **Status** (lines 56-62): `open #2563EB`, `fixed #64748B`, `archived #64748B`.
- **Verification** (lines 64-70): `unverified #47569`, `community_verified #15803D`, `admin_verified #166534`.
- **Neutral palette** (lines 72-82): `#FCFDFF`, `#F8FAFC`, `#F1F5F9`, `#FFFFF`, `#E2E8F0`, `#0F172A`, `#64748B`.
- **Dark surfaces** (lines 84-93): `Canvas #0B1119`, `Surface #0F1620`, `Card #1E293B`, `Border #334155`, `Text #E5E7EB`, `Muted #9CA3AF`.
- **Dark adjustments** (lines 95-102): open `#60A5FA`, fixed `#9CA3AF`; shadows `rgba(0,0,0,0.30-0.45)`.
- **Typography** (lines 104-112): family `Plus Jakarta Sans, system-ui, …`; Display `clamp(2rem,4vw,3.2rem)/800`; Page Title `28-38px/800`; Section `20-24px/800`; Body `14-15px/400`; Small `12px`; Kicker `11px uppercase 700`.
- **Spacing scale** (line 116): `4·8·12·16·20·24·32·40` px.
- **Radius** (lines 118-126): Card `24-28px`, Button `16-18px`, Sheet `24px`, Badge `999px`, Input `16px`.
- **Shadow** (lines 128-134): Card `0 14px 34px rgba(15,23,42,0.08)`, Soft `…0.06`, Elevated `0 20px 48px …0.10`, CTA `0 18px 40px rgba(229,72,77,0.22)`, Marker `0 2px 6px rgba(0,0,0.35)`.
- **Layout container** (lines 136-142): public narow `560px`, public wide `1200px`, admin `1320px`.
- **Component primitive class names** (lines 144-154): `jedug-card`, `jedug-card-soft`, `jedug-panel`, `admin-card`, `section-title`, `section-kicker`, `surface-label`, `metric-card`, `state-panel`, `error-panel`, `notice-panel`.
- **Form/button rules** (lines 156-172): input min 48px; primary `#E5484D` white text, min 48px, radius `16-18px`; secondary white bg/`#E2E8F0` border; icon button square 44px radius 16px.
- **Icon rules** (lines 174-182): sizes `16/18/20-24px`; family Solar line-duotone via `frontend/src/lib/icons.ts`.
- **Marker spec** (lines 184-189): touch 36×36, dots `14/16/20px`, border 2.5px white, selected `scale(1.5)`.
- **Bottom sheet** (lines 192-198): handle 40×4px `#CBD5E1`, top radius 20px, mobile max 55vh, desktop side panel 380px.
- **Button spec** (lines 200-206): primary `#E5484D` `12px radius` min-height 48px; active `scale(0.97)`; hover opacity 0.88. **NOTE: contradicts §"Form & Button Rules" which says radius 16-18px.**
- **Z-index scale** (lines 210-219): `0/5/8/10/15/20/1000`.

### `/Users/ryanprayoga/Project/Pribadi/jedug_v2/design-docs/component-spec.md` (360 lines)

Component-level behavior specs. Defines:

- AppHeader (16-47), NearbyAlertsPanel (49-68), IssueMap (70-106), Marker (108-115), Report Location Panel (117-127), IssueBottomSheet (129-148), IssueCard (150-161), IssueHeader (163-174), IssueStats (176-184), IssueFollowCard (186-221), StatsDashboard (223-235), AdminLogin (237-250), AdminShell & Moderation (252-260), Notification-Driven Issue Refresh (262-268), NotificationPreferencesPanel (270-280), IssueGallery (282-289), ShareActions (291-304), Issue Detail Activity (306-313), PrimaryButton (315-322), EmptyState (324-330), ErrorState (332-336), LoadingState (338-342), ConsentSheet (344-350).
- Concrete values quoted: PrimaryButton `#E5484D` 12px radius min 52px (line 318), EmptyState text `#64748B` (line 328), ErrorState text `#DC2626` (line 335), LoadingState track `#E2E8F0` top `#E5484D` (line 341), ConsentSheet sheet `20px top radius max 480px pading 24px 20px` (line 347).

### `/Users/ryanprayoga/Project/Pribadi/jedug_v2/design-docs/guide.md` (322 lines)

Page-level UX guidance + change log. Defines:

- Global layout/font/runtime (34-40), AppHeader polish (42-47), IssueMap markers/heatmap (49-55), IssueBottomSheet (57-63), PrimaryButton/IssueCard/EmptyState/ErrorState/LoadingState/ConsentSheet/ImagePicker (65-99).
- Page sections: Landing `/` (104-109), Peta `/issues` (111-134), Stats `/stats` (136-170), Lapor `/lapor` (172-184), Header nav (186-214), Detail `/issues/[id]` (216-269), Login admin `/admin/login` (271-281), Admin moderation `/admin/*` (283-289).
- Migration table "Sebelum→Sesudah" (lines 295-306) lists historical color maping.
- Trailing line 322: `TEST DEPLOY` (stray junk in source-of-truth doc).

---

## 2. `frontend/src/app.css` (456 lines) — Defined Surface

### Tailwind theme tokens (`@theme` block, lines 3-54)
- `--font-sans` = `"Plus Jakarta Sans", ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif` (lines 4-11).
- Brand scale 50-900: `#ff4f3` `#ffe3e1` `#ffc9c6` `#ff9e98` `#f56e67` `#e5484d` `#c5363a` `#a1282b` `#7f1f22` `#5c181a` (lines 14-23).
- Surface scale 25-900: `#fcfdff` `#f8fafc` `#f1f5f9` `#e2e8f0` `#cbd5e1` `#94a3b8` `#64748b` `#334155` `#0f172a` (lines 26-34).
- Severity 1/2/3/5: `#f6c453` `#f97316` `#dc2626` `#991b1b` (lines 37-40). **Severity 4 missing.**
- Status: `--color-status-open #2563eb`, `--color-status-fixed #64748b`, `--color-status-archived #64748b` (lines 43-45).
- Verification: `--color-verification-unverified #475569`, `--color-verification-community #15803d`, `--color-verification-admin #166534` (lines 46-48).
- Shadow tokens: `--shadow-card 0 14px 34px rgba(15,23,42,0.08)`, `--shadow-soft 0 20px 48px rgba(15,23,42,0.10)`, `--shadow-brand 0 18px 40px rgba(229,72,77,0.22)` (lines 51-53).

**Missing tokens vs design-system.md:** No spacing scale, no radius scale, no typography scale, no z-index scale, no heatmap palette, no dark-mode-specific surface tokens (`#0B1119`, `#0F1620`, `#1E293B`, `#9CA3AF`, etc.), no severity-4 token.

### `@font-face` / `@import`
- Line 1: `@import "tailwindcss";` (only stylesheet import).
- **No `@font-face`.** Plus Jakarta Sans loaded via `<link>` in `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/app.html:9-12` (Google Fonts CDN).

### Base/reset rules (`@layer base`, lines 56-188)
- Lines 57-63: `:root { color-scheme: light }`, `:root.dark { color-scheme: dark }`.
- Lines 65-69: universal `box-sizing: border-box`.
- Lines 71-78: `html { scroll-behavior: smooth; background: #f8fafc }`, `html.dark { background: #0b1119 }`. **Hardcoded hex bypassing tokens.**
- Lines 80-98: `body` font-family via token, color `#0f172a` (hardcoded), background uses three radial/linear gradients with hardcoded `rgba(229,72,77,0.11)`, `rgba(251,191,36,0.10)`, `#fcfdff`, `#f8fafc`, `#f1f5f9`. Dark variant lines 92-98 hardcoded `#0f1620` `#0b1119` `#070d15`.
- Lines 100-102: `a { color: inherit }`.
- Lines 104-109: `button, input, textarea, select { font: inherit }`.
- Lines 111-119: `::selection` light `rgba(229,72,77,0.18)/#7f1f22`, dark `rgba(229,72,77,0.35)/#fca5a5`.
- Lines 121-140: `:focus-visible` ring stack with hardcoded rgba.
- Lines 142-187: MapLibre control overides (`.maplibregl-ctrl-*`) with hardcoded rgba borders/backgrounds, `border-radius: 1rem !important` (line 155), button size `44px` (lines 169-170).

### Components (`@layer components`, lines 190-438)

Full enumeration:

| Class | Lines | Purpose |
|---|---|---|
| `.app-shell` | 191-193 | flex column min-h-dvh |
| `.app-main` | 195-197 | max-w-560 px-4 pb-8 pt-6 |
| `.app-main-wide` | 199-201 | max-w-1200 |
| `.app-main-full` | 203-205 | max-w-none, no padding |
| `.public-stack` | 207-209 | flex col gap-5 |
| `.jedug-card` + dark | 211-217 | radius 28px, white, shadow |
| `.jedug-card-soft` + dark | 219-225 | radius 24px, blur backdrop |
| `.jedug-panel` + dark | 227-233 | radius 24px, slate panel |
| `.section-kicker` + dark | 235-241 | pill kicker brand |
| `.section-title` + dark | 243-249 | clamp display 800 |
| `.section-copy` + dark | 251-257 | body slate-600 |
| `.surface-label` + dark | 259-265 | uppercase tracking-0.18em |
| `.input-shell` | 267-269 | flex col gap-2 |
| `.input-label` + dark | 271-277 | semibold slate-700 |
| `.input-help` + dark | 279-285 | xs slate-500 |
| `.input-field` + dark | 287-293 | h-12 rounded-2xl |
| `.textarea-field` + dark | 295-301 | min-h-28 rounded-3xl |
| `.select-field` + dark | 303-309 | h-12 rounded-2xl |
| `.btn-primary` | 311-313 | min-h-12 brand-500 shadow brand |
| `.btn-secondary` + dark | 315-321 | white border slate |
| `.btn-ghost` + dark | 323-329 | min-h-11 transparent |
| `.btn-danger` + dark | 331-337 | rose tints |
| `.btn-icon` + dark | 339-345 | size-11 square |
| `.badge-muted` + dark | 347-353 | pill slate |
| `.badge-tint` + dark | 355-361 | pill brand |
| `.metric-card` + dark | 363-369 | radius 22px |
| `.metric-label` + dark | 371-377 | tracking-0.16em |
| `.metric-value` + dark | 379-385 | text-2xl 800 |
| `.metric-copy` + dark | 387-393 | xs slate-500 |
| `.state-panel` + dark | 395-401 | radius 24px center |
| `.error-panel` + dark | 403-409 | rose 22px |
| `.notice-panel` + dark | 411-417 | amber 22px |
| `.admin-shell-bg` + dark | 419-425 | linear gradient bg |
| `.admin-frame` | 427-429 | max-w-1320 |
| `.admin-card` + dark | 431-437 | radius 28px |

### Utilities (`@layer utilities`, lines 440-455)
- `.text-balance` (lines 441-443): `text-wrap: balance`.
- `.surface-ring` + dark (lines 445-455): double box-shadow ring.

---

## 3. Theme Store — `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/stores/theme.ts` (98 lines)

**API:**
- Type `ThemeMode = 'light' | 'dark' | 'system'` (line 4).
- Persist key: `'jedug-theme'` in `localStorage` (line 6).
- Default: `'system'` (line 41 fallback in `getInitialTheme`).
- Reads system preference via `window.matchMedia('(prefers-color-scheme: dark)')` (lines 8-11).

**Functions:**
- `getInitialTheme()` (lines 40-42): returns stored or `'system'`.
- `setTheme(theme)` (lines 44-53): writes `localStorage`, cals `applyTheme`.
- `getResolvedTheme()` (lines 55-58): returns `'light'|'dark'`.
- `useThemeSync()` (lines 60-77): subscribes to OS `change` event; only re-applies when stored is `'system'`. Returns cleanup.
- `applyTheme(theme)` (lines 24-38): sets `data-theme` attribute AND toggles `.dark`/`.light` classes on `document.documentElement`. **Both attribute and class used; design-system.md only mentions `data-theme`.**

**Object export `theme`** (lines 80-96): `current` geter, `resolved` getter, `set(value)`, `getIcon()` returning `'sun'|'moon'`, `getLabel()` returning `'Mode Terang'|'Mode Gelap'`.

**Reactivity gap:** `theme.current`/`theme.resolved` are plain getters, NOT Svelte stores or `$state`. They do not auto-trigger re-renders when changed via `setTheme`. Coment line 79 says "Svelte 5 store-like API" — misleading.

**Consumers:**
- `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/routes/+layout.svelte:12,28,40` — initial aply + sync hook.
- `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/components/AppHeader.svelte:17,44` — toggle UI.

---

## 4. Inline Styles & `<style>` Blocks

### Inline `style=` attribute usage in Svelte (only 5 sites)
- `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/components/IssueBottomSheet.svelte:158` — `style="transform: translateY({dragOffsetY}px);"` — drag offset (legitimate dynamic).
- `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/components/IssueBottomSheet.svelte:174` — `style={`background: ${severityColor[…]}`}` — severity color literal injection (TOKEN LEAK; bypasses CSS vars).
- `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/components/IssueBottomSheet.svelte:180` — `style={getStatusStyle(issue.status)}` — status bg/text from JS literals (TOKEN LEAK).
- `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/components/IssueCard.svelte:45` — `style={`background: ${severityColor[…] || '#94A3B8'}`}` — TOKEN LEAK.
- `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/components/IssueHeader.svelte:109,112,115` — three inline `background:`/`color:` from JS tone maps (TOKEN LEAK).
- `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/routes/issues/[id]/+page.svelte:916` — `style={`background:${getSeverityColor(submission.severity)}`}` — TOKEN LEAK.
- `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/routes/stats/+page.svelte:819` — `style={`width:${item.percent}%`}` — legitimate dynamic.

### `<style>` blocks in `.svelte` files (only 3)
- `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/components/IssueMap.svelte:870-921` — map loading overlay/card/dot. Hardcoded `#47569`, `#e5484d`, all rgba literals.
- `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/components/IssueBottomShet.svelte:276-332` — sheet animations + responsive split. No tokens used.
- `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/components/ConsentSheet.svelte:25-92` — entire component styled outside Tailwind. Hardcoded `#ff`, `#0F172A`, `#64748B`, `#E5484D`. `font-size: 18px/14px/16px`, `font-weight: 700/600`, `border-radius: 20px 0 0`/`12px`, `pading: 24px 20px`, `pading: 14px`.

**Verdict:** ConsentSheet is the worst offender — completely untokenized. IssueHeader/IssueBottomSheet/IssueCard/IssueMap inject hex via inline style.

---

## 5. Color Usage — Top Offenders (hardcoded hex / rgba bypassing tokens)

Top sites where code injects color literals instead of using tokens. Full count: **88 hex matches** across 13 files; **40 rgba matches** across 8 files (excluding `app.css` self-definition).

| # | Path:line | Literal | Context |
|---|---|
| 1 | `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/utils/issue-detail.ts:32-39` | `#EF6FF #2563EB #F1F5F9 #64748B #FEF2 #DC2626 #F8FAFC #F0FDF4 #16A34A #FEF3C7 #B45309` | `STATUS_TONES` map — entirely untokenized status palette |
| 2 | `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/utils/issue-detail.ts:52-57` | `#DCFCE7 #166534 #ECFDF5 #15803D #F1F5F9 #47569 #FEF3C7 #92400E #FEE2E2 #991B1B` | `VERIFICATION_TONES` map — untokenized |
| 3 | `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/utils/issue-detail.ts:61,88` | `', '#F6C453','#F97316','#DC2626','#DC2626','#991B1B'`; fallback `#94A3B8` | `SEVERITY_COLORS` array (duplicates token) |
| 4 | `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/components/IssueMap.svelte:112-119` | `#94A3B8 #991B1B #DC2626 #F97316 #F6C453` | `markerColorExpression` MapLibre literals |
| 5 | `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/components/IssueMap.svelte:145-151` | `#FCA5A5 #FB7185 #E5484D #BE123C` | `clusterCircleColorExpression` cluster ramp (NOT in design-system) |
| 6 | `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/components/IssueMap.svelte:398,414-415,442,453,469` | `#FFFF #FFFF #7F1D1D #FFFF #E5484D #FFFF` | layer paints (cluster stroke, count text/halo, marker, selected core) |
| 7 | `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/components/IssueCard.svelte:12,45` | `#F6C453 … #94A3B8` | duplicated severity array |
| 8 | `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/components/IssueBottomSheet.svelte:36,174` | `#F6C453 … #94A3B8` | duplicated severity array |
| 9 | `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/components/ConsentSheet.svelte:52,56,79` | `#0F172A #64748B #E5484D` | scoped `<style>` |
| 10 | `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/components/IssueMap.svelte:894,905` | `#47569 #e5484d` | scoped `<style>` |
| 11 | `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/components/IssueHeader.svelte:93` | `#1e293b #0f172a` + `rgba(229,72,77,0.18)` | hero placeholder gradient via Tailwind arbitrary value |
| 12 | `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/routes/issues/+page.svelte:208,434` | `#fcfdff #f8fafc #ef2f7`; `#f8fafc #ef2ff` | full-page gradients via arbitrary tailwind |
| 13 | `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/routes/issues/+page.svelte:417` | `rgba(246,196,83,0.45) rgba(249,115,22,0.72) rgba(229,72,77,0.92) rgba(153,27,27,0.98)` | inline heatmap legend gradient (DOES match design-system heatmap palette but as raw rgba) |
| 14 | `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/routes/api/og/issues/[id]/+server.ts:14,174` | `#E5484D #B91C1C #7F1D1D` | OG image generator (server) |
| 15 | `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/routes/admin/login/+page.svelte:68,171` | `rgba(229,72,77,0.16) #ff8f7 #ff #e5484d` | hero gradient + checkbox accent |

Other sites: `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/routes/stats/+page.svelte:579` `#fff9f8 #ffff`; `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/routes/lapor/+page.svelte:597` `#e5484d`; `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/routes/issues/[id]/+page.svelte:114` `#94A3B8`; `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/components/NotificationPreferencesPanel.svelte:191` `#e5484d`; `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/components/NearbyAlertsPanel.svelte:283` `#e5484d`; `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/routes/+layout.svelte:90` `rgba(15,23,42,0.14)`.

---

## 6. Typography — Hardcoded literals not behind a CSS variable

Total: **`font-family` 3 hits, `font-size` 10 hits, `font-weight` 9 hits.** Of those, only `app.css:82 font-family: var(--font-sans)` uses a variable.

| Path:line | Literal | Context |
|---|---|
| `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/components/IssueMap.svelte:895` | `font-size: 0.75rem` | `.map-loading-card` |
| `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/components/IssueMap.svelte:896` | `font-weight: 700` | same |
| `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/components/ConsentSheet.svelte:50` | `font-size: 18px` | `h2` |
| `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/components/ConsentSheet.svelte:51` | `font-weight: 700` | `h2` |
| `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/components/ConsentSheet.svelte:55` | `font-size: 14px` | `.consent-body` |
| `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/components/ConsentSheet.svelte:76` | `font-size: 16px` | `.accept-btn` |
| `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/components/ConsentSheet.svelte:77` | `font-weight: 600` | `.accept-btn` |
| `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/routes/api/og/issues/[id]/+server.ts:138` | `font-size:64px;font-weight:700;font-family:Arial,sans-serif` | OG fallback |
| `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/routes/api/og/issues/[id]/+server.ts:188` | `font-family:'Segoe UI','Inter',Arial,sans-serif` | OG main |
| `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/routes/api/og/issues/[id]/+server.ts:197` | `font-size:26px;font-weight:700` | OG kicker |
| `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/routes/api/og/issues/[id]/+server.ts:201` | `font-size:48px;font-weight:800;line-height:1.14` | OG title |
| `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/routes/api/og/issues/[id]/+server.ts:204` | `font-size:36px;font-weight:600` | OG sub |
| `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/routes/api/og/issues/[id]/+server.ts:207` | `font-size:30px;font-weight:500` | OG meta |
| `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/routes/api/og/issues/[id]/+server.ts:211` | `font-size:30px;font-weight:800` | OG badge |

**Tailwind text-* utilities throughout components (text-xs/sm/[15px]/2xl/clamp(...)) bypass the doc's typography scale silently — none of `28-38px`, `20-24px`, `12px`, `11px` exist as named tokens.** Only `clamp(1.65rem,4vw,2.9rem)` appears (`app.css:244` `.section-title`) and even that diverges from the doc's `clamp(2rem,4vw,3.2rem)`.

---

## 7. Spacing / Radius — Raw values not behind a variable (top 15)

Pure `pading`/`margin`/`border-radius`/`gap` literals (CSS): 6 `border-radius` hits, 10 pading/margin/gap hits in scoped styles. Tailwind arbitrary values `rounded-[Npx]`: **129 occurrences across 22 files** — every page uses ad-hoc radii.

| Path:line | Literal | Context |
|---|---|---|
| `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/components/ConsentSheet.svelte:38` | `border-radius: 20px 0 0` | `.consent-sheet` |
| `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/components/ConsentSheet.svelte:81` | `border-radius: 12px` | `.accept-btn` |
| `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/components/ConsentSheet.svelte:39` | `pading: 24px 20px` | sheet |
| `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/components/ConsentSheet.svelte:75` | `padding: 14px` | btn |
| `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/components/ConsentSheet.svelte:49,61` | `margin: 0 12px`, `0 0 8px` | h2/p |
| `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/components/IssueBottomSheet.svelte:320` | `pading: 12px 0 4px` | `.sheet-handle-area` |
| `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/components/IssueMap.svelte:879,888,891` | `padding: 1rem`, `gap: 0.75rem`, `pading: 0.875rem 1rem` | overlay/card |
| `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/components/IssueMap.svelte:890,904` | `border-radius: 999px` | pill/dot |
| `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/app.css:155` | `border-radius: 1rem !important` | maplibre ctrl |
| `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/routes/stats/+page.svelte:618,629,642` | `rounded-[22px]` ×3 | stat cards |
| `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/routes/stats/+page.svelte:802,846,931` | `rounded-[24px]` ×3 | leaderboard cards |
| `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/routes/issues/+page.svelte:231,282,295,306,324,339,347,372,390,409` | `rounded-[22px] rounded-[20px] rounded-[24px] rounded-[34px]` | tolbar/cards/badges |
| `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/routes/issues/[id]/+page.svelte:819,826,831,911,988,1069` | `rounded-[22px] rounded-[24px] rounded-[30px]` | various cards |
| `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/routes/+layout.svelte:90` | `rounded-[24px]` | toast |
| `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/routes/admin/login/+page.svelte:67` | `rounded-[34px]` | login shell |

**Doc claims radius scale `24-28/16-18/24/999/16`. Code uses `18, 20, 22, 24, 28, 30, 34` arbitrarily.** No spacing tokens at all in `app.css`; design-system.md spacing scale `4·8·12·16·20·24·32·40` not encoded anywhere.

---

## 8. Iconography — `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/icons.ts` (44 lines)

All from `@iconify-svelte/solar` library, family `line-duotone`. **No inline SVG.** 43 named exports:

`AddCircleIcon, ClockIcon, ArowLeftIcon, ArowRightIcon, ArrowUpIcon, BellIcon, CameraIcon, ChartIcon, CheckCircleIcon, ClipboardIcon, CloseCircleIcon, CompassIcon, CopyIcon, DangerIcon, DocumentIcon, EyeClosedIcon, EyeIcon, FilterIcon, GaleryIcon, HamburgerIcon, HistoryIcon, InfoIcon, LayersIcon, ListCheckIcon, LocationIcon, LoginIcon, LogoutIcon, MapIcon, MonIcon, RankingIcon, NotificationIcon, RefreshIcon, RouteIcon, SearchIcon, SettingsIcon, ShareIcon, ShieldCheckIcon, StretsIcon, SunIcon, TargetIcon, TrashIcon, UserIcon, UsersGroupIcon, WidgetIcon`

Sources map exactly: `solar/<slug>-line-duotone`. Aligns with design-system.md icon rule.

Only non-iconify SVG aset in repo: `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/assets/favicon.svg` (favicon only, not icon system).

---

## Critical Gap Summary (for redesign brief)

1. **No spacing/radius/typography tokens in CSS.** Doc scale exists only on paper; code uses Tailwind arbitrary values everywhere (`rounded-[22px]`, `text-[15px]`).
2. **Severity-4 token missing** in `app.css` despite scale 1/2/3/5 defined.
3. **Status & verification tones live in JS** (`/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/utils/issue-detail.ts:32-57`) — completely outside CSS layer, injected via inline `style=`.
4. **Severity color array duplicated 4×** (`issue-detail.ts:61`, `IssueCard.svelte:12`, `IssueBottomSheet.svelte:36`, `IssueMap.svelte:112-119`).
5. **ConsentSheet entirely untokenized** scoped CSS — easiest single-file fix.
6. **Dark-mode surface tokens (`#0B119`, `#0F1620`, `#1E293B`) NOT in `app.css`**; hardcoded inline at lines 77, 97, 420, 424.
7. **Heatmap palette (4 rgba steps) not tokenized**; reproduced literally at `/Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/routes/issues/+page.svelte:417`.
8. **Cluster colors (`#FCA5A5 #FB7185 #E5484D #BE123C`) not in design-system at all** — exists only in code (`IssueMap.svelte:145-151`).
9. **Theme store `theme.current/resolved` are not reactive** — geters not `$state`. UI listening to it won't auto-rerender.
10. **Internal contradiction in design-system.md**: button radius `16-18px` (line 164) vs `12px` (line 202).
11. **Stray `TEST DEPLOY` line** at `design-docs/guide.md:322` pollutes source-of-truth.
12. **Font loaded via Google `<link>` not `@font-face`/self-host** — performance + privacy gap.

<results>
<files>
- /Users/ryanprayoga/Project/Pribadi/jedug_v2/design-docs/design-system.md - token source-of-truth (palette, type scale, spacing, radius, shadow, z-index)
- /Users/ryanprayoga/Project/Pribadi/jedug_v2/design-docs/component-spec.md - component behavior/state specs with concrete hex/size values
- /Users/ryanprayoga/Project/Pribadi/jedug_v2/design-docs/guide.md - page-level UX guidance + historical migration notes (contains stray "TEST DEPLOY")
- /Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/app.css - actual Tailwind v4 @theme tokens + base/components/utilities; missing spacing/radius/type scales
- /Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/app.html - Plus Jakarta Sans loaded via Google Fonts <link>
- /Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/stores/theme.ts - theme store; localStorage key `jedug-theme`, default `system`; NOT reactive
- /Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/icons.ts - 43 Iconify Solar line-duotone exports (no inline SVG)
- /Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/utils/issue-detail.ts - status/verification/severity tone maps in JS bypassing CSS tokens
- /Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/components/ConsentSheet.svelte - fully untokenized scoped <style>
- /Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/components/IssueMap.svelte - hardcoded marker/cluster/heatmap hex + scoped <style>
- /Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/components/IssueBottomSheet.svelte - inline severity/status style injection + scoped <style>
- /Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/components/IssueCard.svelte - duplicated severity color array, inline style
- /Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/components/IssueHeader.svelte - 3 inline tone style injections
- /Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/routes/issues/+page.svelte - heatmap legend rgba literals, 10+ rounded-[Npx] arbitrary values
- /Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/routes/issues/[id]/+page.svelte - inline severity bg, multiple rounded-[Npx]
- /Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/routes/api/og/issues/[id]/+server.ts - OG image with hardcoded BRAND_RED + Arial/Segoe UI fonts (server context, separate concern)
</files>

<answer>
Design-system docs claim a complete token system (palette, 8-step spacing, 5 radii, 6-step type scale, z-index lader) but `app.css` only encodes color + font-sans + 3 shadow tokens. **No spacing, radius, or typography tokens exist in CSS.** Components compensate with Tailwind arbitrary values (`rounded-[22px]`, `text-[15px]`, 129 occurrences) and JS color maps injected via inline `style=`. Severity color array is duplicated 4 times across files. Status/verification palettes live entirely in `/lib/utils/issue-detail.ts` not CSS. Dark-mode surface tokens (`#0B1119`, `#0F1620`, `#1E293B`) referenced in docs are NOT in `app.css` — they're hardcoded in base layer + admin-shell-bg gradients. Heatmap palette + cluster ramp exist only as raw rgba/hex in code. Theme store works (localStorage `jedug-theme`, modes `light|dark|system`, default `system`, applies both `data-theme` attr AND `.dark` class) but `theme.current/resolved` are plain geters, not Svelte 5 `$state`, so subscribers won't auto-rerender. Icons fully aligned (43 Solar line-duotone exports). Worst untokenized component: ConsentSheet. Worst leakage pattern: status/severity tones via inline `style={`background:${tone.bg}`}`.
</answer>

<next_steps>
For `docs/REDESIGN_AUDIT.md`, recommend the new tokens layer must add:
1. CSS custom properties for spacing (`--space-1..8`), radius (`--radius-sm/md/lg/pill`), type scale (`--text-xs..display`), z-index ladder, dark-mode surface scale, heatmap 4-step rgba, cluster ramp, status/verification tone pairs.
2. Migrate `/lib/utils/issue-detail.ts` STATUS_TONES/VERIFICATION_TONES/SEVERITY_COLORS to reference CSS vars (e.g., `var(--tone-status-open-bg)`).
3. Convert ConsentSheet scoped style to Tailwind utility classes hiting tokens.
4. Convert `theme.current/resolved` to `$state` writable rune so AppHeader auto-syncs.
5. Resolve doc contradiction: button radius single value (recommend 14px or 16px, pick one).
6. Strip `TEST DEPLOY` from `guide.md:322`.
7. Self-host Plus Jakarta Sans via `@font-face` instead of Google `<link>`.
Ready to draft the brief.
</next_steps>
</results>