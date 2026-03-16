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
- Nav links: "Lapor", "Peta", "Statistik", 14px, 500 weight, `#64748B`, hover `#E5484D`
- Padding: 12px 16px
- Notification bell:
  - badge unread tampil dari jumlah item dengan `read_at = null`
  - panel dropdown tetap nyaman di mobile (`<= 340px` lebar efektif)
  - panel menampilkan card ringan `Notifikasi Browser` sebelum daftar item
  - panel juga menampilkan section `Preferensi Notifikasi` yang bisa expand/collapse tanpa pindah halaman
  - section preferensi minimum berisi:
    - master switch
    - toggle channel in-app
    - toggle channel push
    - toggle per event type
  - CTA browser push tidak boleh auto-trigger permission prompt; prompt hanya muncul setelah user tap tombolnya
  - tiap item punya area tap utama untuk membuka issue dan action hapus terpisah yang tetap mudah disentuh
  - action hapus tidak boleh ikut memicu navigasi item
  - item unread memakai tint ringan merah muda agar cepat dibedakan
  - jika browser push belum aktif, toggle push boleh disabled selama ada status/CTA yang jelas untuk mengaktifkannya

## IssueMap

- Full viewport minus header
- Map controls: top-right, compact
- Loading overlay: top-left, pill badge, spinning dot + "Memuat..."
- Map info badge: top-left, pill, menampilkan `{n} titik` + status area (ada/tidak ada laporan)
- Empty state map tidak menggunakan popup tengah; informasi "tidak ada laporan" ditampilkan di info badge
- Error overlay: top full-width, red bg
- Harus punya toggle visual mode:
  - `Marker`
  - `Heatmap`
- Marker publik memakai layer stack:
  - cluster circles
  - cluster count
  - unclustered hit-area
  - unclustered marker dot
  - selected glow/core
- Heatmap publik memakai:
  - density heat layer
  - subtle point accent di zoom lebih dekat
- Toggle mode harus stabil, tanpa flicker add/remove source berulang
- Klik cluster harus zoom/focus ke area cluster
- Saat heatmap aktif:
  - marker individual dan cluster disembunyikan
  - bottom sheet tidak muncul
  - tampilkan legend intensitas ringkas yang tetap aman di mobile
- Jika setup heatmap gagal, fallback otomatis ke marker mode tanpa menjatuhkan map ke mode list
- Jika setup cluster gagal, fallback ke unclustered marker layer (tanpa memutus flow map)

## Marker

- Severity-based colors: Ringan `#F6C453`, Sedang `#F97316`, Berat `#DC2626`
- Fixed/archived: `#94A3B8`, opacity 0.45
- Touch target: 36px
- Visual dot sizes: 14/16/20px by severity
- Selected: glow + core marker (base marker disembunyikan via filter)
- Cluster count harus tetap terbaca di mobile (text + halo)

## Report Location Panel (`/lapor`)

- User selalu melihat koordinat mentah (`lat, lon`) sebagai acuan utama.
- Label lokasi manusiawi (wilayah) ditampilkan sebagai konfirmasi UX, bukan source of truth geospatial.
- Lookup label dipicu saat:
  - lokasi awal berhasil didapat
  - koordinat manual dipilih eksplisit
- Jangan melakukan request berulang saat user sedang mengetik koordinat.
- Jika lookup label gagal, submit tetap bisa lanjut dengan koordinat mentah.
- UI menampilkan nama lokasi sebagai label manusiawi (primary + secondary line) dan bukan field editable.
- Nama jalan di issue dilengkapi backend saat submit report (reverse geocode fallback), jadi panel lokasi cukup fokus ke konfirmasi koordinat + wilayah.

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

## IssueFollowCard

- Ditempatkan di detail issue publik dekat metrik utama agar CTA follow cepat terlihat.
- Isi minimum:
  - judul `Ikuti Perkembangan`
  - follower count (`X orang mengikuti laporan ini`)
  - tombol toggle `Ikuti laporan ini` / `Berhenti mengikuti`
  - helper text anonim-friendly
  - error state ringan bila request gagal
- Tombol:
  - full width
  - min-height 48px
  - default state memakai brand red `#E5484D`
  - state following memakai tinted surface `#FFF1F2` + border `#FECDD3`
  - disabled opacity turun, tidak boleh terlihat broken
- Loading state:
  - tombol disable saat request berlangsung
  - label tombol berubah menjadi state progres
- Copy UX harus menegaskan bahwa satu browser/device anonim dihitung sebagai satu follower.
- Jika browser ini sudah follow issue:
  - card boleh menampilkan sub-card `Aktifkan Notifikasi Browser`
  - state copy minimum:
    - browser tidak mendukung
    - izin ditolak
    - izin granted tapi subscription belum aktif
    - notifikasi browser aktif
    - subscription gagal
- Tombol browser push:
  - min-height 48px
  - tetap memakai brand red `#E5484D`
  - jangan tampil sebagai modal agresif atau interstitial

## Notification-Driven Issue Refresh

- Jika user klik notifikasi dan target issue berbeda dari route aktif, lakukan navigasi normal ke `/issues/[id]`.
- Jika user klik notifikasi ketika sudah berada di `/issues/[id]` yang sama:
  - jangan noop
  - refresh detail issue, timeline, dan state follow/follower count
  - tampilkan micro-feedback ringan bahwa laporan diperbarui

## NotificationPreferencesPanel

- Ditempatkan di notification center, bukan halaman settings baru.
- Default state boleh collapse agar dropdown tetap ringkas.
- Copy harus manusiawi dan anonim-friendly; hindari istilah teknis seperti `event_type`.
- Jika follower/browser belum punya binding notif yang sah:
  - tampilkan helper text yang menjelaskan user perlu follow setidaknya satu issue dulu
  - jangan tampilkan toggle yang terasa broken
- Jika `notifications_enabled = false`:
  - child toggle tetap terlihat
  - child toggle boleh dibuat disabled untuk menegaskan bahwa master switch sedang mematikan semuanya

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
