# JEDUG Design System Specification

## Scope dan Source of Truth

Dokumen ini adalah source of truth untuk:

- visual style global
- token warna/tipografi/spacing/radius/shadow/z-index
- primitive visual yang dipakai lintas komponen

Dokumen ini bukan tempat utama untuk behavior detail tiap komponen atau flow halaman.

- behavior komponen: lihat `design-docs/component-spec.md`
- page-level UX: lihat `design-docs/guide.md`

## Brand Identity

- **Brand Name**: JEDUG
- **Tagline**: Pantau Jalan Rusak
- **Character**: Civic, serious, clean, informative, mobile-first, map-first

## Color Palette

### Brand

| Token     | Hex       | Usage                         |
| --------- | --------- | ----------------------------- |
| JEDUG Red | `#E5484D` | Logo, primary CTA, active nav |

### Severity

| Level | Label  | Hex       | Usage         |
| ----- | ------ | --------- | ------------- |
| 1     | Ringan | `#F6C453` | Marker, badge |
| 2     | Sedang | `#F97316` | Marker, badge |
| 3+    | Berat  | `#DC2626` | Marker, badge |

### Status

| Status   | Label         | Hex       | Usage                      |
| -------- | ------------- | --------- | -------------------------- |
| open     | Terbuka       | `#2563EB` | Status badge               |
| verified | Terverifikasi | `#16A34A` | Status badge               |
| fixed    | Selesai       | `#64748B` | Status badge, muted marker |

### Neutral

| Token          | Hex       | Usage                        |
| -------------- | --------- | ---------------------------- |
| Background     | `#F8FAFC` | Page background              |
| Card / Sheet   | `#FFFFFF` | Cards, sheets, header        |
| Border         | `#E2E8F0` | Dividers, card borders       |
| Text Primary   | `#0F172A` | Headings, body text          |
| Text Secondary | `#64748B` | Captions, labels, muted text |

## Typography

- **Font Family**: `Inter, system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif`
- **Page Title**: 20px / weight 700
- **Section Title**: 16px / weight 600
- **Body**: 14px / weight 400
- **Small / Caption**: 12px / muted color

## Spacing System

`4px · 8px · 12px · 16px · 24px · 32px`

## Corner Radius

| Element      | Radius       |
| ------------ | ------------ |
| Card         | 16px         |
| Button       | 12px         |
| Bottom sheet | 20px         |
| Badge        | 999px (pill) |
| Input        | 10px         |

## Shadow

- **Card**: `0 1px 3px rgba(0,0,0,0.06), 0 1px 2px rgba(0,0,0,0.04)`
- **Elevated (sheet, floating)**: `0 4px 16px rgba(0,0,0,0.10)`
- **Marker**: `0 2px 6px rgba(0,0,0,0.35)`

## Marker Spec

- Touch target: 36×36px
- Dot size: Ringan 14px, Sedang 16px, Berat+ 20px
- Border: 2.5px white
- Selected state: scale(1.5), drop-shadow glow
- Fixed/archived: opacity 0.45

## Bottom Sheet

- Drag handle: 40×4px, radius 2px, `#CBD5E1`
- Border-radius top: 20px
- Shadow: elevated
- Max height mobile: 55vh
- Desktop: side panel 380px, no radius, border-left

## Button Spec

- Primary: `#E5484D` bg, white text, 12px radius, min-height 48px
- Secondary: white bg, `#0F172A` text, `#E2E8F0` border, 12px radius
- Disabled: opacity 0.45, cursor not-allowed
- Active: scale(0.97)
- Hover: opacity 0.88

## Z-Index Scale

| Layer        | Z-Index |
| ------------ | ------- |
| Map base     | 0       |
| Map overlays | 5       |
| Map controls | 8       |
| Bottom CTA   | 10      |
| Side panel   | 15      |
| Bottom sheet | 20      |
| Header       | 100     |
| Consent      | 1000    |

## Read This Next

- `design-docs/component-spec.md`
- `design-docs/guide.md`
- `docs/DESIGN_INDEX.md`
