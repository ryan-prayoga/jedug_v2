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
- **Frontend Styling Runtime**: Tailwind CSS v4 + theme tokens di `frontend/src/app.css`
- **Icon System**: Iconify Solar `line-duotone` sebagai family default lintas publik + admin

## Color Palette

### Brand

| Token     | Hex       | Usage                         |
| --------- | --------- | ----------------------------- |
| JEDUG Red | `#E5484D` | Logo, primary CTA, active nav |
| Red Deep  | `#C5363A` | Hover primary CTA, emphasis   |
| Red Soft  | `#FFF4F3` | Soft accent surface           |

### Severity

| Level | Label  | Hex       | Usage         |
| ----- | ------ | --------- | ------------- |
| 1     | Ringan | `#F6C453` | Marker, badge |
| 2     | Sedang | `#F97316` | Marker, badge |
| 3+    | Berat  | `#DC2626` | Marker, badge |

### Heatmap

| Level  | Color                  | Usage                    |
| ------ | ---------------------- | ------------------------ |
| Low    | `rgba(246,196,83,0.34)` | Outer heat / low density |
| Medium | `rgba(249,115,22,0.56)` | Mid intensity            |
| High   | `rgba(229,72,77,0.82)`  | High intensity           |
| Peak   | `rgba(153,27,27,0.94)`  | Peak hotspot center      |

### Status

| Status   | Label         | Hex       | Usage                      |
| -------- | ------------- | --------- | -------------------------- |
| open     | Terbuka       | `#2563EB` | Status badge               |
| fixed    | Selesai       | `#64748B` | Status badge, muted marker |
| archived | Diarsipkan    | `#64748B` | Status badge, muted marker |

### Verification

| Status             | Label                   | Hex       | Usage                  |
| ------------------ | ----------------------- | --------- | ---------------------- |
| unverified         | Belum diverifikasi      | `#475569` | Verification badge     |
| community_verified | Terverifikasi komunitas | `#15803D` | Verification badge     |
| admin_verified     | Diverifikasi admin      | `#166534` | Verification badge     |

### Neutral

| Token          | Hex       | Usage                        |
| -------------- | --------- | ---------------------------- |
| Canvas Top     | `#FCFDFF` | Hero/light page background   |
| Background     | `#F8FAFC` | Page background              |
| Surface Soft   | `#F1F5F9` | Panel, muted card, subtle bg |
| Card / Sheet   | `#FFFFFF` | Cards, sheets, header        |
| Border         | `#E2E8F0` | Dividers, card borders       |
| Text Primary   | `#0F172A` | Headings, body text          |
| Text Secondary | `#64748B` | Captions, labels, muted text |

## Typography

- **Font Family**: `Plus Jakarta Sans, system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif`
- **Display / Hero**: `clamp(2rem, 4vw, 3.2rem)` / weight `800`
- **Page Title**: `28-38px` / weight `800`
- **Section Title**: `20-24px` / weight `800`
- **Body**: `14-15px` / weight `400`
- **Small / Caption**: `12px` / muted color
- **Kicker / Eyebrow**: `11px`, uppercase, tracking lebar, weight `700`

## Spacing System

`4px · 8px · 12px · 16px · 20px · 24px · 32px · 40px`

## Corner Radius

| Element      | Radius       |
| ------------ | ------------ |
| Card         | 24-28px      |
| Button       | 16-18px      |
| Bottom sheet | 24px         |
| Badge        | 999px (pill) |
| Input        | 16px         |

## Shadow

- **Card**: `0 14px 34px rgba(15,23,42,0.08)`
- **Soft / Glass**: `0 14px 34px rgba(15,23,42,0.06)`
- **Elevated (sheet, floating)**: `0 20px 48px rgba(15,23,42,0.10)`
- **Primary CTA**: `0 18px 40px rgba(229,72,77,0.22)`
- **Marker**: `0 2px 6px rgba(0,0,0,0.35)`

## Container & Layout Rules

- Public narrow content: max `560px`
- Public wide/detail content: max `1200px`
- Admin frame: max `1320px`
- Global cards memakai layout card-based dengan vertical gap `20px`
- Background tidak flat polos; pakai gradient/radial lembut agar UI terasa hidup namun tetap civic-grade

## Component Primitives

- Primary surface: `jedug-card`
- Soft/glass surface: `jedug-card-soft`
- Muted panel: `jedug-panel`
- Admin surface: `admin-card`
- Heading utility: `section-title`
- Kicker utility: `section-kicker`
- Helper/muted label: `surface-label`
- Metric primitive: `metric-card`
- State primitive: `state-panel`, `error-panel`, `notice-panel`

## Form & Button Rules

- Input/select tinggi minimum `48px`
- Form shell wajib memakai label + helper text bila konteks tidak obvious
- Primary button:
  - bg `#E5484D`
  - text putih
  - min-height `48px`
  - radius `16-18px`
- Secondary button:
  - white bg
  - border `#E2E8F0`
  - shadow lembut
- Icon button:
  - square `44px`
  - radius `16px`
  - dipakai untuk close, show/hide password, action sekunder

## Icon Rules

- Gunakan Iconify via `frontend/src/lib/icons.ts`
- Family default: Solar `line-duotone`
- Ukuran umum:
  - inline status/icon label: `16px`
  - button/icon chip: `18px`
  - section marker: `20-24px`
- Hindari campuran icon library lain kecuali ada kebutuhan teknis yang belum ter-cover

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
