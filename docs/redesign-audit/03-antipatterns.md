Task Result

Task ID: bg_9eccc9d7
Description: Audit UI antipatterns and AI-slop
Duration: 6m 15s
Session ID: ses_13cd41aa3ffeCq6683JwpArmqp

---

<analysis>
**Literal Request**: Adversarial audit of frontend/src for paterns fighting minimalist-editorial redesign across 10 categories with file:line citations.

**Actual Need**: Actionable remediation hit-list to fed into redesign brief — needs concrete paths, lines, and prescriptive fixes.

**Success Looks Like**: Markdown doc, H2 per category, each bullet = `path:line — issue — fix`. Real cites, no fluff.
</analysis>

OK, I have enough material. Compose remediation hit-list now.

<results>
<files>
- /Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/app.css - design token base, all surface/button vocabulary
- /Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/routes/+page.svelte - landing hero & quick links
- /Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/routes/lapor/+page.svelte - report form (long, dense)
- /Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/routes/issues/+page.svelte - public map + list
- /Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/routes/issues/[id]/+page.svelte - public issue detail
- /Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/routes/stats/+page.svelte - public stats (961 LOC, dense)
- /Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/routes/admin/+layout.svelte - admin shell
- /Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/routes/admin/login/+page.svelte - admin login
- /Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/routes/admin/issues/+page.svelte - admin moderation list
- /Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/routes/admin/issues/[id]/+page.svelte - admin issue detail + actions
- /Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/components/IssueMap.svelte - MapLibre layers + marker logic
- /Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/components/IssueBottomSheet.svelte - map bottom sheet
- /Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/components/IssueCard.svelte - public issue card
- /Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/components/IssueHeader.svelte - detail page hero
- /Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/components/IssueGallery.svelte - photo grid
- /Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/components/IssueStats.svelte - 4-up metric grid
- /Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/components/AppHeader.svelte - public top bar + notif panel
- /Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/components/ImagePicker.svelte - photo upload box
- /Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/components/EmptyState.svelte - shared empty state
- /Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/components/ErorState.svelte - shared error state
- /Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/components/LoadingState.svelte - shared loading
- /Users/ryanprayoga/Project/Pribadi/jedug_v2/frontend/src/lib/components/ConsentSheet.svelte - first-visit consent (uses scoped CSS, off-system)
</files>

<answer>

# Minimalist-Editorial Remediation Hit-List

Adversarial pass over `frontend/src`. All paths absolute. Skiped tests, TODOs, generated types.

## 1. Generic AI / shadcn-y Paterns

- `frontend/src/app.css:84-87,94-97` — body uses dual radial-gradient + linear-gradient ambient bg (warm pink + amber blob) → flaten to single warm off-white (`#FAF7` / `#F8F5F0`) or pure paper-white; remove all radial blobs.
- `frontend/src/app.css:212,220,228,396,432` — every "card" macro stacks `rounded-[28px]/[24px]` + `shadow-[0_14px_34px_...]` + soft white border + sometimes `backdrop-blur` → kill the floating-card vocabulary; replace with hairline `border border-stone-200` + `rounded-md` (or none) + zero shadow. Magazine bento sits flat on page.
- `frontend/src/routes/+page.svelte:82,86,90` — three identical "Lapor → Pantau → Tindak lanjuti" mini-cards each `rounded-[22px] border border-white/80 bg-white shadow-[0_10px_24px...]` nested inside `jedug-panel` already nested inside `jedug-card` (3 levels of bordered surfaces) → replace with numbered editorial list `01 / 02 / 03` over single hairline divider, no nested cards.
- `frontend/src/routes/+page.svelte:104-105` — primary CTA `'btn-primary min-h-[76px] justify-start rounded-[26px] px-5'` is a giant pill with brand red gradient shadow → convert to text-link with arow: `Laporkan jalan rusak →`, weight 600, single underline-on-hover. Save brand red for severity, not nav.
- `frontend/src/routes/+page.svelte:69-72` — four pill badges (`Mobile-first`, `Anonim-friendly`, `Map-first`, `Moderation-ready`) are decorative buzwords → delete; replace with one editorial dek line.
- `frontend/src/routes/+page.svelte:99-124` — three quick-link cards repeat icon-square + bold + caption shadcn-card pattern → editorial: numbered serif list with role labels (`Lapor`, `Peta`, `Statistik`), no card chrome.
- `frontend/src/lib/components/EmptyState.svelte:6,16-25` — `icon = '📭'` emoji default + brand-tinted icon square; magazine-grade UIs avoid emoji decoration → drop emoji prop entirely; use thin rule + small caps label.
- `frontend/src/lib/components/IssueHeader.svelte:79` — `bg-gradient-to-t from-slate-950 via-slate-950/25 to-transparent` overlay on hero photo (instagram-card pattern) → replace with bottom leterbox black bar + serif caption, or no overlay (caption below image like editorial print).
- `frontend/src/lib/components/IssueHeader.svelte:93` — empty-photo state uses `radial-gradient + linear-gradient(135deg, #1e293b → #0f172a)` dark gradient placeholder → replace with flat warm beige box, single line "Belum ada foto" in small caps.
- `frontend/src/routes/admin/login/+page.svelte:68` — left panel `bg-[radial-gradient(circle_at_top_left,rgba(229,72,77,0.16),...)]` marketing-hero gradient → flat off-white split, hairline divider between hero + form.
- `frontend/src/routes/issues/+page.svelte:208` — page wrapper `bg-[radial-gradient(...)+linear-gradient(...)]` → flat surface.
- `frontend/src/routes/+page.svelte:104,109-110` vs `frontend/src/lib/components/IssueCard.svelte:58` vs `frontend/src/lib/components/IssueBottomSheet.svelte:192` — same "icon square in brand-tinted rounded-[18px]/2xl" stamp duplicated 12+ times → kill the tinted icon-square; use plain inline svg in body color.

## 2. Trust-Hostile Paterns (Civic Tech)

- `frontend/src/lib/components/IssueCard.svelte:81,88,97,104` — stat labels `text-[11px] uppercase tracking-[0.16em]` (LAPORAN/FOTO/KORBAN/STATUS) — at 11px + tight tracking on mobile this is borderline illegible, especially in dark mode → bump to 12px, lower-case roman, weight 500. Reserve all-caps for one global kicker.
- `frontend/src/lib/components/IssueCard.svelte:62` vs `82,90,99` — h-tier ambiguity: location label is `text-base font-bold` (16px/700), stat values are `text-lg font-[800]` (18px/800) → stat numbers visually outweigh title. Editorial fix: title 20-24px serif/semibold, stats 14px tabular-nums, no 800-weight numbers.
- `frontend/src/routes/+page.svelte:104` `min-h-[76px]` for CTA is fine, but `frontend/src/lib/components/AppHeader.svelte:325-333` notification delete button uses `size-11` (44px) wrapped in li with arbitrary spacing — verify compound list-item still hits 44px target on first paint.
- `frontend/src/lib/components/IssueCard.svelte:43-53` — severity + status both encoded ONLY by background color (red/blue/emerald/rose pills) → add icon glyph or text-prefix (`Severity 4 — Parah`); never rely on color alone.
- `frontend/src/lib/components/IssueMap.svelte:105-120` — markers encode severity by color only (`#991B1B → #F6C453` ramp) with no shape or size differentiation for fixed/archived (only opacity 0.45) → add stroke pattern (dashed for fixed) or marker shape (circle vs ring) so colorblind users distinguish.
- `frontend/src/lib/components/IssueMap.svelte` (entire) — no on-map legend; legend conditional only in heatmap (`showHeatmapLegend`, `routes/issues/+page.svelte:45`). Marker mode has zero legend → add tiny editorial legend strip (4 swatches + label) bottom-left, persistent.
- `frontend/src/app.css:121-140` — focus-visible exists globally but is a soft red halo `0 0 0 4px rgba(229,72,77,0.28)` on white — contrast against light brand backgrounds is weak → switch to 2px solid `#0F172A` ring + 2px white offset for A on any surface.
- `frontend/src/app.css:71-72` `scroll-behavior: smooth` on html — disables jump-link teleport for users with motion sensitivity → wrap with `@media (prefers-reduced-motion: no-preference)`.
- `frontend/src/lib/components/LoadingState.svelte:6` + `frontend/src/lib/components/IssueBottomSheet.svelte:301-308` + `frontend/src/lib/components/AppHeader.svelte:244` — spinners + slide animations on data-critical screns, no `prefers-reduced-motion` guard → wrap all `animate-spin`, `animation: slideUp/slideInRight`, `map-loading-pulse` keyframes.
- `frontend/src/lib/components/IssueMap.svelte:909-919` — `@keyframes map-loading-pulse` infinite pulse on data fetch; runs forever until viewport ready → reduced-motion variant should be static dot.

## 3. Heavy-Shadow / Heavy-Gradient (count by file)

- `frontend/src/app.css` — 15 occurrences (4 hero gradients, 11 shadow tokens) → strip every shadow token (`--shadow-card/soft/brand`), strip body radial blobs, strip `surface-ring` utility.
- `frontend/src/routes/issues/+page.svelte` — 6 `backdrop-blur` + 11 `shadow-[…]` + 3 gradient bg → flatten to white sections, hairline borders only.
- `frontend/src/routes/issues/[id]/+page.svelte` — 9 `shadow-[…]` (most on info-cards already inside `.jedug-card`) → kill nested shadows; one wrapper, no iner cards.
- `frontend/src/lib/components/AppHeader.svelte` — 4 `backdrop-blur` + 6 shadows including `shadow-[0_14px_30px_rgba(229,72,77,0.2)]` brand-red glow on active nav (`:354-356`) → flat 1px bottom border, active link = bottom underline only, no glow.
- `frontend/src/lib/components/IssueBottomSheet.svelte:155` — `bg-white/96 shadow-[0_-12px_40px_rgba(15,23,42,0.18)] backdrop-blur-xl` floating glassmorphism sheet → solid white sheet, single 1px top border, no blur.
- `frontend/src/lib/components/IssueBottomShet.svelte:264` overlay = `linear-gradient(180deg, rgba(15,23,42,0.04), rgba(15,23,42,0.18))` → flat single dim `rgba(15,23,42,0.32)`.
- `frontend/src/routes/issues/[id]/+page.svelte:1061,1069,1072` preview lightbox stacks `bg-slate-950/82 backdrop-blur-sm` + `shadow-[0_30px_80px...]` + `border border-white/10 bg-slate-950` + `bg-slate-900/82 backdrop-blur` (close btn) → solid black overlay, no blur, no shadow on iner frame.
- `frontend/src/lib/components/ImagePicker.svelte:32,47,51` — gradient on dropzone bg + gradient overlay on preview + `inset shadow + scale-105 hover` → flat bordered dropzone, dashed `border-stone-300`, no hover-grow.
- `frontend/src/routes/api/og/issues/[id]/+server.ts:3` (3 gradient hits) — OG image gen uses linear/radial; OK to keep since social previews benefit from contrast, but match new palette.
- `frontend/src/routes/admin/+layout.svelte:55,98` admin-shell-bg uses `bg-[linear-gradient(180deg,#f8fafc,#ef2ff)]` purple tint → flat warm-grey paper background.

## 4. Layout Antipatterns (cards-in-cards-in-cards)

- `frontend/src/routes/+page.svelte:54 → 76 → 82/86/90` — `.jedug-card` (level 1) → `.jedug-panel` (level 2) → 3 iner `rounded-[22px] border bg-white shadow-[…]` (level 3). Three levels of bordered/rounded surfaces → collapse to single editorial layout with type hierarchy.
- `frontend/src/routes/admin/issues/[id]/+page.svelte:137 → 198 → 213/224 → individual buttons` — `.admin-card` → `.jedug-panel` → moderation form with each button + input wrapped in own border. 3+ levels.
- `frontend/src/routes/admin/issues/[id]/+page.svelte:254 → 257-275` — `.admin-card.p-5` containg 4 `rounded-[22px] border bg-slate-50` metadata tiles each with own border (already inside metric-cards box) → flat list with `dt/dd` editorial table.
- `frontend/src/routes/issues/+page.svelte:210 → 282/295/306` — `.jedug-card-soft` wrapping 3 mini-cards each with own border + shadow + tints → flaten the tolbar, mode summary should be one inline strip.
- `frontend/src/routes/lapor/+page.svelte:390 → 426 → 441 → 449` — main hero card → location section card → emerald confirmation card → badge inside it. 4 levels.
- `frontend/src/routes/lapor/+page.svelte:611 → 596 → input row` — `.jedug-card` wraps casualty form which wraps `.rounded-[22px] border` checkbox row → only one border layer need.
- `frontend/src/lib/components/IssueBottomSheet.svelte:155 → 190 → 192 (icon-bg-square)` — sheet → location card → icon tile. Reduce one level.
- Magic max-widths repeated:
  - `560px` (`app-main` `app.css:196`)
  - `1200px` (`app-main-wide` `app.css:200`, `AppHeader.svelte:171`)
  - `1320px` (`admin-frame` `app.css:428`)
  - `1180px`, `1080px`, `420px` ad-hoc in `routes/issues/+page.svelte:480`, `routes/admin/login/+page.svelte:67`, `routes/issues/+page.svelte:559` → consolidate to two tokens (`max-w-prose 65ch` for editorial copy, `max-w-content 1200px` for grid).

## 5. Density Problems

- `frontend/src/lib/components/IssueCard.svelte:77-107` — 4 stat tiles in `grid-cols-2 sm:grid-cols-4` each `px-3 py-3` with bordered chrome and AL-CAPS labels — too cramped at mobile width and visually too loud → drop chrome, render as 4 inline `value · label` pairs separated by small bullets, line-height 1.5.
- `frontend/src/routes/issues/+page.svelte:281-316` — 3 viewport summary articles each with full chrome (border + shadow + bg-white/90) repeating mode/viewport/action info already shown in tolbar → delete entirely or reduce to single status line.
- `frontend/src/routes/+page.svelte:99-138` — two separate `grid gap-3` sections (quickLinks + valueCards) showing 6 cards total on landing — too sparse, scans like blog fed not navigation → merge into single editorial block.
- `frontend/src/app.css:380` `metric-value` is `text-2xl font-[800] tracking-[-0.03em]` — 24px/800 is screaming, with all-caps 11px label above and 12px copy below: density imbalance → metric-value 28-32px serif lining-figures, label 11px sentence-case below, no all-caps.
- `frontend/src/app.css:196` `max-w-[560px]` for `.app-main` forces narow column on desktop (good for mobile), but combined with `space-y-5` (20px) and 28px-radius cards → too sparse on tablet/desktop. Use 720px reading column at md+, 1080-1200 grid at lg+.
- `frontend/src/lib/components/IssueBottomSheet.svelte:209-239` — 4-tile grid in the bottom sheet duplicates IssueCard's 4-tile grid — high density mobile → reduce to 2 stats max in sheet (Severity + Reports), rest in detail page.

## 6. Inconsistent Component Vocabulary (duplications)

- **Severity pill** implemented 4 ways:
  - `frontend/src/lib/components/IssueCard.svelte:43-48` inline with `severityColor[]` array literal
  - `frontend/src/lib/components/IssueBottomSheet.svelte:172-177` same array inline duplicated
  - `frontend/src/lib/components/IssueHeader.svelte:109-111` via `getSeverityColor()` util
  - `frontend/src/routes/admin/issues/[id]/+page.svelte:151,149` admin renders raw `detail.status` text without same pill
  → Extract `<SeverityPill severity={n} />` and `<StatusPill status={s} />` components, source-of-truth color in one util.
- **Status tone map** duplicated:
  - `frontend/src/lib/components/IssueCard.svelte:20-29` `statusTone` Record literal
  - `frontend/src/routes/admin/issues/+page.svelte:47-60` `statusTone()` function (different palette: emerald for open, vs blue/slate in card)
  - `frontend/src/routes/admin/issues/[id]/+page.svelte:102-115` same function copy-pasted
  - `frontend/src/lib/utils/issue-detail.ts` exports `getStatusTone` returning `{bg,text}` style object — yet 3 call sites use class-based palette → consolidate on class-based `<StatusPill>` reading single map.
- **Stat tile** ("Laporan/Foto/Korban") implemented 3+ ways:
  - `frontend/src/lib/components/IssueCard.svelte:78-105`
  - `frontend/src/lib/components/IssueBottomSheet.svelte:210-239`
  - `frontend/src/lib/components/IssueStats.svelte:50-63` (uses `.metric-card`)
  - `frontend/src/routes/admin/issues/[id]/+page.svelte:163-195` (uses `.metric-card`)
  → Single `<MetricTile label value caption alert />` component.
- **Spinner** implemented 4 ways:
  - `frontend/src/lib/components/LoadingState.svelte:6` `size-11 border-[3px] border-slate-200 border-t-brand-500`
  - `frontend/src/routes/admin/+layout.svelte:106` identical inline
  - `frontend/src/routes/admin/issues/+page.svelte:118` identical
  - `frontend/src/routes/admin/issues/[id]/+page.svelte:120` identical
  - `frontend/src/lib/components/AppHeader.svelte:244,275` `size-9 border-[3px]` (different size)
  - `frontend/src/routes/admin/login/+page.svelte:182` `size-4 border-2 border-white/35` (different size + color)
  → Use `<LoadingState>` everywhere or extract `<Spinner size>`.
- **Card classes**: `.jedug-card` / `.jedug-card-soft` / `.jedug-panel` / `.metric-card` / `.state-panel` / `.admin-card` (6 variants in `app.css`) — most differ only by radius (22-28px) and shadow opacity → collapse to 2 (`Card` + `Card.muted`) for editorial language.

## 7. Map / Marker UX

- `frontend/src/lib/components/IssueMap.svelte:105-120` — color-only severity encoding (already flaged above).
- `frontend/src/lib/components/IssueMap.svelte` entire — no permanent legend in marker mode; only conditional heatmap legend rendered from `routes/issues/+page.svelte:45` → add fixed mini-legend (sev1-5 swatches + fixed-state ring) as inline editorial caption bottom-left.
- `frontend/src/lib/components/IssueMap.svelte:873-883` `.map-loading-overlay` uses radial gradient bg → flat warm overlay.
- `frontend/src/lib/components/IssueMap.svelte:885-899` `.map-loading-card` is pill-shaped with shadow + uppercase tracking — yet another pill vocabulary → minimal text label "Memuat peta…" with thin rule.
- `frontend/src/lib/components/IssueBottomSheet.svelte:155 + 287-294` — sheet `max-height: 62vh` covers majority of viewport on tall phones, fights map gesture; drag-to-close threshold `96px` is reasonable but only `.sheet-handle-area` triggers if user touches inside scroled content (`canStartDrag:74` blocks if `scrollTop>0`) — fine, but no escape via tap-on-overlay when sheet is large → already implemented (overlay click), but verify on iOS safari momentum scroll. Also: no error/empty/loading state for the sheet itself if issue is null — only `{#if visible && issue}` short-circuits, never renders skeleton.
- `frontend/src/routes/issues/+page.svelte:42-43` map loading state covers whole map but shows nothing when `issues.length>0` and refetch is in flight → add subtle "memuat ulang" strip top of map.
- `frontend/src/routes/issues/+page.svelte:480-489` map error fallback uses `.notice-panel` (amber) but blocks the whole list with full-width baner → make dismissible inline.
- `frontend/src/lib/components/IssueBottomSheet.svelte:201-204` — popup typography is bold 16px label + 12px cords + bullet — inconsistent with editorial body scale; and `text-slate-500` 12px on white = ~A marginal → bump 13px stone-600.

## 8. Form UX

- `frontend/src/routes/lapor/+page.svelte:626` `placeholder="Deskripsi singkat kondisi jalan (opsional)"` — full sentence in placeholder, no separate label visible above → has `input-label` "Catan tambahan" further up, but the textarea itself only labeled by section heading; placeholder still doubles as instruction → make placeholder short ("Tambahkan konteks…") and ensure persistent helper text outside input.
- `frontend/src/routes/lapor/+page.svelte:503,512` `placeholder="Latitude, mis. -6.2000"` — placeholder-as-label antipattern; no `<label>` association → wrap with `.input-shell` like login form does.
- `frontend/src/lib/components/NearbyAlertsPanel.svelte:193,208,218` — three placeholder-as-label inputs (`Mis. Rumah / Kantor`, `-6.200000`, `106.81666`) → same fix.
- `frontend/src/routes/lapor/+page.svelte:597,605` — checkbox + number-input for casualty has zero `required` indicator and no inline validation copy on bad ranges → add `*` red glyph next to required fields and inline error region with `aria-live="polite"`.
- `frontend/src/routes/admin/login/+page.svelte:133,150` — only place using HTML `required` attribute (2 hits in entire codebase) → all other forms rely on JS-side validation, missing semantic `required` for scren readers.
- `frontend/src/app.css:288` `.input-field` h-12 (48px), `.select-field:304` h-12 → consistent. But `frontend/src/routes/admin/issues/+page.svelte:103` ads `pl-12` for icon, while `routes/admin/login/+page.svelte:131,148` also uses `pl-12 pr-12` — icon-pading spec is ad-hoc, not tokenized → token `.input-field--with-leading-icon`.
- `frontend/src/routes/lapor/+page.svelte:635-643` — submit progress + error are separate cards stacked — error scrolled into view via `errorRef.scrollIntoView` but no `aria-live` or `role="alert"` → `<div role="alert" aria-live="assertive">` for `.error-panel`.
- `frontend/src/lib/components/AppHeader.svelte:74-83` `formatNotifTime` uses `id-ID` locale — fine, but no `<time datetime>` semantic element used anywhere in lists.
- `frontend/src/routes/lapor/+page.svelte:561-578` — severity radios hide `<input class="hidden">` and rely on visual radio-card only; no `aria-checked` exposed on label, no keyboard focus ring on the card → use `radiogroup` pattern with visible focus on the `<label>`.

## 9. Admin Dashboard

- `frontend/src/routes/admin/issues/+page.svelte:132-186` — table has no `<thead>` sticky positioning; on long lists header scrolls away → `thead { position: sticky; top: 0 }` + bg-white.
- `frontend/src/routes/admin/issues/+page.svelte:179-181` — "Buka detail" is the only action, but no inline hide/fix buttons; admins must click into detail for every action → add row action menu OR kep current but document.
- `frontend/src/routes/admin/issues/[id]/+page.svelte:231,242` — destructive actions (`Sembunyikan`, `Tolak issue`) execute immediately on click without confirm. Only `handleBanDevice:66` has `confirm()` → add inline confirm or modal for hide/reject; reason input is "Opsional" placeholder but should be required for moderation audit trust.
- `frontend/src/routes/admin/issues/+page.svelte:47-60` — status tones: `open=emerald`, `fixed=blue`, `rejected=rose`, `archived=slate`. But `frontend/src/lib/components/IssueCard.svelte:21` uses `open=blue`, `verified=emerald`. Same status, different color across views → operationally confusing for moderators fliping between admin and public.
- `frontend/src/routes/admin/issues/+page.svelte:124-130` — empty state present (good). No empty-state for filter combinations (e.g. status=rejected returns 0): same generic copy. → tailor copy by filter context.
- `frontend/src/routes/admin/issues/+page.svelte:191` — pagination shows "Halaman X" but no total count, no items-on-page → add `Menampilkan 1–20 dari N`.
- `frontend/src/routes/admin/issues/+page.svelte:178-181` — action button is `btn-secondary min-h-10 px-4 py-2` (40px) — below 44px tap target on mobile (admin still needs that). Same in `:193,197` Prev/Next.
- `frontend/src/routes/admin/+layout.svelte:69-94` — admin top bar has redundant chrome: nav pill + user pill + logout button each in own bordered shadow container → flaten to single bar with text links and inline logout.
- `frontend/src/routes/admin/issues/[id]/+page.svelte:286` — evidence photos open in new tab via `<a href={public_url} target="_blank">` — no inline lightbox like public detail. Inconsistent with public UX.
- `frontend/src/routes/admin/+page.svelte:1-10` — admin index just renders `<p>Mengarahkan ke daftar issue...</p>` then JS goto. No fallback if JS disabled. Minor.

## 10. Accessibility Inventory

- `aria-` usage: only **25 hits in 12 files** — coverage gaps:
  - `frontend/src/lib/components/IssueCard.svelte` (entire) — 0 aria attrs; severity pill, stat tiles, link card all rely on visual + text but no `aria-label` clarifies "Buka detail issue X".
  - `frontend/src/lib/components/IssueStats.svelte` — has section-level `aria-label`, but each `<article>` metric is decorative without role.
  - `frontend/src/routes/admin/issues/+page.svelte` — 0 aria attrs; table rows have no `scope="col"` on `<th>`.
  - `frontend/src/routes/admin/issues/[id]/+page.svelte` — 0 aria attrs except via icons; destructive buttons lack `aria-describedby` linking to the reason field.
  - `frontend/src/routes/issues/+page.svelte:232` has one `aria-label="Mode visual peta"` on toggle group, good — but toggles inside lack `aria-pressed`.
  - `frontend/src/routes/issues/[id]/+page.svelte:1062` `role="button" tabindex="0"` on `<div>` for preview overlay — anti-pattern; use `<button>` or close on Esc only.
- Color contrast obvious failures:
  - `frontend/src/app.css:260` `.surface-label text-slate-400` (#94A3B8) on white — 3.4:1 → fails A for body. Aceptable for "decorative" small caps but used widely as label.
  - `frontend/src/app.css:372` `.metric-label text-slate-400` same issue at 11px → bump to slate-500.
  - `frontend/src/lib/components/IssueCard.svelte:81/88/97/104` `text-slate-500` icon + `text-slate-500` 11px label = ~4.6:1 borderline at 11px, fails WCAG large-text rules for sub-14px.
  - `frontend/src/lib/components/IssueHeader.svelte:99` `text-slate-200` body on `bg-[#0f172a]` → passes, but `:82` `text-white/80` over photo with bottom-25%-opacity gradient overlay can fall below A on bright photos.
  - `frontend/src/routes/lapor/+page.svelte:444` `text-emerald-700` on `bg-emerald-50/80` — passes at 14px but the surounding 11px `tracking-[0.18em]` label is borderline.
  - Brand button red: `frontend/src/app.css:312` `bg-brand-500 #E5484D` + `text-white` = 4.4:1 → marginal A for body, ok for ≥18px bold which is the case.
- `frontend/src/lib/components/IssueBottomSheet.svelte:151,154` `<div class="sheet-overlay" onclick=...>` — div with click handler, ignored a11y warnings. Should be `<button>` overlay or add `role="dialog"` `aria-modal="true"` to sheet.
- `frontend/src/lib/components/AppHeader.svelte:209` notif button has `aria-haspopup="dialog"` `aria-expanded` — good. But the panel itself (`:222`) is a `<div>`, not `role="dialog"`, no `aria-labelledby`.
- `frontend/src/lib/components/IssueGallery.svelte:53` gallery buttons have `aria-label` (good). Same for hero (`IssueHeader.svelte:69`).
- Missing skip-link: `frontend/src/routes/+layout.svelte:83` no `<a href="#main">Skip to content</a>` for keyboard users.

</answer>

<next_steps>
This list is the "what to remove / what to add" fedstock. Recommended sequence for redesign brief:

1. Strip global tokens first: `app.css:51-53` shadow tokens, `:84-87,94-97` body gradients, `.jedug-card*` shadow stacks → forces every page to fall back to flat surfaces.
2. Consolidate component vocabulary: extract `<StatusPill>`, `<SeverityPill>`, `<MetricTile>`, `<Spinner>` (kills 4-way dupes flaged in §6).
3. Page rewrites in order of impact: `+page.svelte` (landing) → `IssueCard` (used 5+ places) → `IssueHeader` (detail hero) → `IssueBottomSheet` (map) → admin tables.
4. Final passes: a11y (§10), color-contrast token sweep (replace `slate-400` body labels with `stone-500`), motion-reduce wraps.
5. Magazine type stack proposal: serif display (e.g. Source Serif / Cormorant) for h1/h2, kep Plus Jakarta for UI, lining-figures tabular-nums for metrics.

Ready to fed into redesign brief.
</next_steps>
</results>