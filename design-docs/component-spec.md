# JEDUG UI Component Spec

## Scope dan Source of Truth

Dokumen ini adalah source of truth untuk:

- struktur dan perilaku komponen UI
- state visual komponen (normal/hover/selected/disabled)
- aturan interaksi komponen lintas halaman

Dokumen ini bukan sumber utama token desain.

- token visual global: lihat `design-docs/design-system.md`
- guidance page-level UX: lihat `design-docs/guide.md`

## AppHeader

- Sticky top, white bg, bottom border `#E2E8F0`
- Logo: "JEDUG" in brand red `#E5484D`, Inter 700, 1.25rem
- Sub-label: "Pantau Jalan Rusak" in `#64748B`, 11px, next to logo
- Nav links: "Lapor" & "Peta", 14px, 500 weight, `#64748B`, hover `#E5484D`
- Padding: 12px 16px

## IssueMap

- Full viewport minus header
- Map controls: top-right, compact
- Loading overlay: top-left, pill badge, spinning dot + "Memuat..."
- Map info badge: top-left, pill, menampilkan `{n} titik` + status area (ada/tidak ada laporan)
- Empty state map tidak menggunakan popup tengah; informasi "tidak ada laporan" ditampilkan di info badge
- Error overlay: top full-width, red bg

## Marker

- Severity-based colors: Ringan `#F6C453`, Sedang `#F97316`, Berat `#DC2626`
- Fixed/archived: `#94A3B8`, opacity 0.45
- Touch target: 36px
- Visual dot sizes: 14/16/20px by severity
- Selected: scale(1.5), glow shadow
- Hover: scale(1.25)

## IssueBottomSheet

- Mobile: slides from bottom, 20px top radius, drag handle, max-height 55vh
- Desktop: side panel 380px, right-aligned, border-left
- Drag handle: 40×4px, `#CBD5E1`, centered
- Swipe-to-close (mobile):
  - drag down dari handle/area atas sheet
  - close jika melewati threshold gesture
  - jika drag pendek, sheet snap-back ke posisi awal
  - tidak boleh bentrok dengan scroll konten internal
- Content padding: 16px 20px 24px
- Layout:
  1. Severity pill (most prominent, colored bg)
  2. Status pill (separate, clear color coding)
  3. Location (road_name bold, 15px)
  4. Stats row (grid: Laporan, Foto, Korban, Terakhir)
  5. Actions: "Lihat Detail" primary, "Lapor di Sini" secondary
- Severity badge: pill, colored bg, white text, 12px font
- Status badge: pill, colored bg + text according to status color
- Stats: icon-less, value bold, label 11px uppercase muted

## IssueCard

- White bg, `#E2E8F0` border, 16px radius
- Padding: 16px
- Hover: subtle shadow
- Severity pill + status pill at top
- Road name bold
- Meta line: submission count, casualty, relative time

## IssueHeader

- Hero media tampil penuh di mobile, rounded card di desktop
- Jika hero image gagal atau tidak ada, tampilkan placeholder civic-grade yang tetap menonjolkan lokasi issue
- Summary card wajib memuat:
  - severity badge
  - status badge
  - verification badge
  - lokasi utama
  - first seen
  - last seen
  - snapshot singkat issue (`severity · laporan · foto · korban/reaksi jika ada`)

## IssueStats

- Grid 2 kolom di mobile, 5 kolom di desktop
- Item wajib:
  - laporan
  - foto
  - korban
  - reaksi
  - update terakhir
- Card korban boleh diberi state alert ringan jika `casualty_count > 0`

## IssueGallery

- Grid sederhana, tidak memakai carousel
- Klik gambar membuka preview full-screen/lightbox ringan
- Jika total foto lebih besar dari foto yang dikirim endpoint, tampilkan helper text yang menjelaskan bahwa yang tampil adalah subset terbaru
- Empty state harus membedakan:
  - benar-benar belum ada foto
  - foto ada tetapi gagal dimuat di perangkat

## ShareActions

- Harus menyediakan CTA:
  - kembali ke peta
  - bagikan issue
  - lapor di sekitar sini
  - buka lokasi di peta eksternal
- Share links cepat:
  - WhatsApp
  - Telegram
  - Twitter/X
  - Facebook
  - salin link
- Primary share menggunakan Web Share API bila tersedia, lalu fallback ke copy link

## Issue Detail Activity

- Ringkasan aktivitas publik hanya boleh memakai data aman:
  - severity
  - reported_at
  - `public_note` yang sudah diringkas
  - casualty indicator/count jika aman
- Jangan tampilkan moderation note, device info, atau catatan internal lain.

## PrimaryButton

- Full width, `#E5484D`, white text, 12px radius
- Min height: 52px
- Font: 15–16px, 700 weight
- Hover: opacity halus + soft elevated shadow
- Active: scale(0.97)
- Disabled: opacity 0.45

## EmptyState

- Centered vertical flex
- Icon: 48px (emoji)
- Text: 14px, `#64748B`
- Optional CTA link below
- Padding: 48px 16px

## ErrorState

- Same layout as EmptyState
- Red tint: text `#DC2626`
- Retry button: outline red, 12px radius

## LoadingState

- Centered flex
- Spinner: 32px, `#E2E8F0` border, `#E5484D` top-border, spinning
- Text: 14px, `#64748B`

## ConsentSheet

- Fixed overlay, semi-transparent bg
- Sheet: 20px top radius, max-width 480px, padding 24px 20px
- Title: 20px, 700
- Body: 14px, `#64748B`
- Accept button: PrimaryButton style

## Catatan Scope

Spesifikasi halaman seperti landing/report dipindahkan ke `design-docs/guide.md` agar tidak overlap dengan spesifikasi komponen.

## Read This Next

- `design-docs/design-system.md`
- `design-docs/guide.md`
- `docs/DESIGN_INDEX.md`
