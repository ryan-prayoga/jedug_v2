# 01 — Surface Inventory

Semua surface user-facing yang harus diredesain. Tidak ada yang dilewati.

## Public Routes

| Route | File | Primary Task |
|---|---|---|
| `/` | `frontend/src/routes/+page.svelte` | Landing — hero + 3 quick links + value cards |
| `/lapor` | `frontend/src/routes/lapor/+page.svelte` | Submit anonymous road-damage report |
| `/issues` | `frontend/src/routes/issues/+page.svelte` | Public map (default) + list fallback |
| `/issues/[id]` | `frontend/src/routes/issues/[id]/+page.svelte` | Issue detail (hero/gallery/stats/follow/share/timeline) |
| `/stats` | `frontend/src/routes/stats/+page.svelte` | Public statistics dashboard scoped by province/regency |

Plus 1 dynamic image: `/api/og/issues/[id]` — OG share card.

## Admin Routes

| Route | File | Primary Task |
|---|---|---|
| `/admin/login` | `frontend/src/routes/admin/login/+page.svelte` | Admin sign-in (split hero) |
| `/admin` | `frontend/src/routes/admin/+page.svelte` | Redirect ke `/admin/issues` |
| `/admin/issues` | `frontend/src/routes/admin/issues/+page.svelte` | Issue moderation table + filter + pagination |
| `/admin/issues/[id]` | `frontend/src/routes/admin/issues/[id]/+page.svelte` | Issue detail + moderation actions (hide/unhide/fix/reject/ban) |

## Layouts

- `frontend/src/routes/+layout.svelte` — public root shell (AppHeader + ConsentSheet + theme bootstrap).
- `frontend/src/routes/admin/+layout.svelte` — admin chrome (header + nav pill + signed-in card + logout).

**No global footer.** Both shells header-only.

## Reusable Components

### Surface (high impact)

- `AppHeader.svelte` — top nav + brand + notif bell + theme toggle. Used in every public page.
- `ConsentSheet.svelte` — first-visit consent overlay. **Fully untokenized scoped CSS — needs full rewrite.**
- `IssueMap.svelte` — MapLibre integration (markers, clusters, heatmap, basemap switch).
- `IssueBottomSheet.svelte` — mobile sheet for marker tap, drag-to-close.

### Card / Display (high duplication)

- `IssueCard.svelte` — issue summary card.
- `IssueHeader.svelte` — detail page hero + meta block.
- `IssueGallery.svelte` — photo grid + count badge.
- `IssueStats.svelte` — 4-metric strip.
- `ShareActions.svelte` — share buttons + back + external map.
- `EmptyState.svelte` — empty placeholder + CTA.
- `ErrorState.svelte` — error + retry.
- `LoadingState.svelte` — spinner panel.
- `PrimaryButton.svelte` — thin wrapper, **candidate to retire** (most pages call `.btn-primary` class direct).

### Forms / Input

- `ImagePicker.svelte` — photo capture/preview tile.
- `BrowserPushCard.svelte` — web-push enable/disable.
- `NearbyAlertsPanel.svelte` — nearby alert subscriptions CRUD.
- `NotificationPreferencesPanel.svelte` — per-channel toggles.

## Forms (per route)

### Public

- **Lapor** (`routes/lapor/+page.svelte`):
  - photo (`ImagePicker`)
  - latitude/longitude (auto-geolocate + manual override)
  - reverse-geocoded location label confirmation
  - severity radio 1–5 (Ringan/Sedang/Berat/Parah/Kritis)
  - `hasCasualty` checkbox + `casualtyCount` number
  - free-text `note` textarea (max 500)
  - multi-step state: `getting-location` → `compressing` → `preparing-upload` → `uploading` → `submitting` → `done`

- **Nearby alert** (`NearbyAlertsPanel.svelte`):
  - `label` text
  - `latitude/longitude` numeric (with "use my location")
  - `radius` 100–5000m

- **Notification prefs** (`NotificationPreferencesPanel.svelte`):
  - 5 boolean toggles in 3 sections: Umum, Channel, Jenis update

- **Stats filter** (`routes/stats/+page.svelte`):
  - province + regency `<select>` pair

### Admin

- **Login** (`routes/admin/login/+page.svelte`):
  - `username`, `password` (show/hide toggle), `rememberMe`

- **Issue list filter** (`routes/admin/issues/+page.svelte`):
  - `statusFilter` select (open/fixed/rejected/archived)
  - pagination

- **Moderation** (`routes/admin/issues/[id]/+page.svelte`):
  - `reasonInput` textarea
  - 5 action buttons: Hide, Unhide, Mark Fixed, Reject, Ban Device (per submission row)

## Stores Driving UI

- `theme.ts` — light/dark/system, persist key `jedug-theme`. **Currently NOT reactive** (plain getters, not `$state`).
- `notifications.ts` — list + SSE realtime + unread count → AppHeader bell.
- `browser-push.ts` — push subscription state → BrowserPushCard.
- `notification-preferences.ts` — per-channel toggles.
- `nearby-alerts.ts` — subscriptions store.

## Map Pieces

- `IssueMap.svelte` (sole MapLibre integration):
  - 2 sources: `jedug-issues-source` (clustered), `jedug-heatmap-source` (unclustered)
  - 8 layers: cluster circles + counts, unclustered hit/base, selected glow/core, heatmap density + points
  - basemap switches via `getResolvedTheme` between Carto positron + dark-matter
  - marker color expression based on status + severity
  - auto geolocate on first load
- `bbox.ts` — viewport-bounded fetch debounce 300ms.
- `issue-heatmap.ts` — heat weight per issue.
- `geolocation.ts` — browser GPS.

**No marker icon SVG/PNG** — markers are MapLibre circle layers driven by data expressions.

Marker palette today:
- gray `#94A3B8` (fixed/archived)
- red shades `#991B1B` / `#DC2626` / `#F97316` / `#F6C453` (severity descending)

## Static Assets

- `frontend/static/favicon.svg` — JEDUG logo: red rounded square + white pin + gold dot + red road glyph.
- `frontend/static/push-icon.svg` — red bell.
- `frontend/static/og/issue-fallback.svg` — OG fallback (slate gradient + brand red).
- `frontend/src/lib/assets/favicon.svg` — duplicate of root favicon.

**No mockups, illustrations, or photo assets.** Design system today lives di CSS tokens + Iconify Solar duotone.

## Image-Gen Manifest Surfaces (downstream)

Setiap surface ini → minimal 1 prompt mobile + 1 prompt desktop:

1. Landing `/` — minimalist editorial hero
2. Lapor `/lapor` — mobile-first multi-section form
3. Map `/issues` (marker mode) — full-bleed map with floating UI
4. Map `/issues` (heatmap mode + sheet open)
5. Issues list fallback — card grid
6. Issue detail `/issues/[id]` — split above-fold + below-fold
7. Photo preview overlay (modal state)
8. Stats `/stats` — summary + filters + breakdowns + top issues
9. Notification popover (header bell open) — combined push + nearby + prefs
10. Consent sheet first-visit overlay
11. Admin login `/admin/login` — split hero
12. Admin issue list `/admin/issues` — table + filter
13. Admin issue detail `/admin/issues/[id]` — moderation toolbar + submissions table + log
14. Empty / Error / Loading state panel set
15. Map markers / clusters / heatmap legend specimen
16. OG share card variant
17. Favicon + push icon refresh
18. Component library board (button + input + card + sheet + tag system)
