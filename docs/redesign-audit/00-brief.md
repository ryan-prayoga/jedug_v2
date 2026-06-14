# 00 — Redesign Brief (Minimalist Editorial)

## North Star

JEDUG harus terasa seperti **publikasi sipil yang serius**, bukan dashboard SaaS atau template marketing.

Referensi taste:

- Editorial print — The New York Times, Bloomberg Originals, MIT Tech Review.
- Civic-tech serius — gov.uk, 18F, Code for America.
- Magazine layout — generous whitespace, hairline rules, lining-figures untuk angka.

## 5 Prinsip

1. **Flat sebelum dekorasi.** Tidak ada shadow lebih dari 1px hairline; tidak ada gradient di permukaan UI (kecuali peta).
2. **Tipografi memimpin hierarki.** Ukuran + weight + spacing yang menentukan struktur, bukan border atau warna.
3. **Warna untuk arti.** Brand red khusus untuk severity & alert. Status pakai tone netral. Tidak ada warna dekoratif.
4. **Editorial numbering.** `01 / 02 / 03` over hairline divider, bukan card grid.
5. **Reduce motion by default.** Animasi cuma untuk state change critical (toast, sheet open). Semua wrapped `prefers-reduced-motion`.

## Palette Target

### Light

- Paper: `#FAF7F2` (warm off-white, bukan blue-tinted)
- Surface: `#FFFFFF` (true white untuk bento blocks)
- Hairline: `#E8E2D8` (warm border, bukan slate)
- Ink: `#1A1A1A` (near-black, bukan slate-900)
- Muted: `#6B6B6B` (warm gray)
- Subtle: `#9A9A9A` (caption gray)
- Brand Red: `#C5363A` (deeper, less saturated than `#E5484D`)
- Severity ramp: `#F0B847` `#E0732B` `#C5363A` `#A1282B` `#7F1F22`
- Status: open=ink/`#1A1A1A`, fixed=muted/`#6B6B6B`, archived=subtle/`#9A9A9A`
- Verification: ink-stroke (unverified) → community=`#3F6B3F` → admin=`#1F4D1F`

### Dark

- Paper: `#0E0D0B` (warm near-black)
- Surface: `#161513`
- Hairline: `#2A2724`
- Ink: `#F0EBE3`
- Muted: `#9A958C`
- Subtle: `#6B665E`
- Brand Red: `#FF6B6F` (lifted for contrast)
- Severity ramp: `#FFCB6B` `#FF8A4F` `#FF6B6F` `#FF8A8E` `#FFB3B5`

## Typography

- **Display (h1, hero)** — Source Serif 4, 48–72px, weight 600, tracking -0.02em.
- **Page Title (h2)** — Source Serif 4, 32–40px, weight 600, tracking -0.01em.
- **Section (h3)** — Plus Jakarta Sans, 18–22px, weight 600, tracking 0.
- **Body** — Plus Jakarta Sans, 15–17px, weight 400, line-height 1.6.
- **Small** — Plus Jakarta Sans, 13px, weight 500.
- **Caption** — Plus Jakarta Sans, 12px, weight 500, color muted.
- **Kicker** — Plus Jakarta Sans, 11px, weight 600, uppercase, tracking 0.18em. **Hanya 1 per surface.**
- **Numerals** — `font-variant-numeric: lining-nums tabular-nums` untuk semua angka di metric.

## Spacing Scale

`4 · 8 · 12 · 16 · 24 · 32 · 48 · 64 · 96` px. Semua spacing harus pakai token `--space-*`.

## Radius Scale

- `--radius-sm: 4px` — input, small button
- `--radius-md: 8px` — card hairline-only
- `--radius-lg: 12px` — sheet, modal
- `--radius-pill: 999px` — pill (severity, kicker)

**No more `rounded-[28px]` arbitrary values.**

## Shadow

**Dilarang.** Hanya boleh:

- `--shadow-hairline: inset 0 0 0 1px var(--color-hairline)` (ini bukan shadow, ini border).
- `--shadow-focus: 0 0 0 2px var(--color-ink), 0 0 0 4px var(--color-paper)` (focus ring only).

Semua `--shadow-card`, `--shadow-soft`, `--shadow-brand` di `app.css` **dihapus**.

## Anti-patterns Dilarang

1. Gradient di body atau card backdrop.
2. Brand-red glow shadow di nav active state.
3. `backdrop-blur` di sheet/modal (gunakan solid `var(--color-surface)`).
4. Nested cards >2 level (jedug-card → jedug-panel → inner-card = forbidden).
5. Tinted icon-square (icon dalam pill berwarna brand) sebagai dekorasi nav.
6. Emoji di EmptyState atau stat label.
7. `text-slate-400` (3.4:1 contrast) untuk body text di surface terang.
8. Color-only severity encoding (harus + label/glyph/shape).
9. All-caps tracking kicker lebih dari 1× per halaman.
10. Animasi infinite tanpa `@media (prefers-reduced-motion: no-preference)` guard.
