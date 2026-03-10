# JEDUG UI Polish — Panduan & Ringkasan

## Scope dan Source of Truth

Dokumen ini menjadi source of truth untuk:

- guidance UX level halaman
- ringkasan keputusan polish UI lintas halaman
- konteks implementasi visual agar konsisten antar route

Dokumen ini bukan sumber utama token global atau spec detail komponen:

- token global: `design-docs/design-system.md`
- behavior komponen: `design-docs/component-spec.md`

## Apa yang Sudah Dikerjakan

Seluruh frontend JEDUG (Svelte 5 + SvelteKit 2) telah di-polish mengikuti design system yang konsisten dan civic-grade. Tidak ada redesign — hanya penyelarasan warna, tipografi, spacing, dan komponen.

---

## File Design Reference

| File                            | Isi                                                               |
| ------------------------------- | ----------------------------------------------------------------- |
| `design-docs/design-system.md`  | Token lengkap: warna, tipografi, spacing, radius, shadow, z-index |
| `design-docs/component-spec.md` | Spesifikasi per komponen UI                                       |
| `design-docs/guide.md`          | File ini — page-level UX guidance + ringkasan perubahan           |

---

## Komponen yang Di-polish

### Global (`+layout.svelte`)

- Font family: Inter (Google Fonts)
- Background: `#F8FAFC`, text: `#0F172A`, base size: `14px`
- Init toast styling diselaraskan

### AppHeader

- Brand color `#E5484D`, sub-label "Pantau Jalan Rusak"
- Nav link hover: `#FEF2F2` bg tint, padding `6px 12px`
- Sub-label hidden di viewport < 360px

### IssueMap (Marker)

- Severity colors: Ringan `#F6C453`, Sedang `#F97316`, Berat `#DC2626`
- Touch target 36px, dot sizes 14/16/20px
- Selected: `scale(1.5)` + red glow; Fixed: opacity `0.45`

### IssueBottomSheet

- 20px top radius, drag handle 40×4px `#CBD5E1`
- Desktop: 380px side panel; Mobile: 55vh max
- Status badges color-coded (Open=biru, Verified=hijau, Fixed=abu)
- Buttons min-height 48px, primary `#E5484D`

### PrimaryButton

- Background `#E5484D`, min-height 48px
- Disabled opacity `0.45`, active `scale(0.97)`

### IssueCard

- Severity & status color-coded badges
- Border-radius 16px, font 12px badges

### EmptyState

- Props baru: `ctaHref` + `ctaLabel` untuk tombol CTA opsional
- Padding 48px, button style `#E5484D`

### ErrorState

- Warna `#DC2626`, bg `#FEF2F2`, border `#FECACA`
- Radius 12px

### LoadingState

- Spinner `#E5484D`, track `#E2E8F0`, padding 48px

### ConsentSheet

- Heading 18px/700, body `#64748B`
- Button `#E5484D`, min-height 48px, `scale(0.97)` active

### ImagePicker

- Placeholder bg `#F8FAFC`, border `#E2E8F0` dashed
- Hover `#F1F5F9` + `#CBD5E1` border
- Change label 12px/500

---

## Halaman yang Di-polish

### Landing (`/`)

- Hero diperkuat dengan kicker/trust statement ringan + card treatment lembut
- Hierarki teks diperjelas: brand title tegas, tagline lebih readable, intro lebih terstruktur
- CTA utama/sekunder diselaraskan (radius 12px, min-height ~52px, shadow lembut)

### Peta & Daftar Laporan (`/issues`)

- Map overlay 10px radius, side panel 360px
- Empty state dengan icon + CTA link ke /lapor
- Toolbar, badges, filter semua ke design spec
- Bottom CTA `#E5484D`, 48px min-height
- Guard state transisi list ↔ map untuk menghindari false-empty/flicker saat map remount
- Empty state map hanya tampil setelah fetch viewport valid, bukan saat map baru mount
- Bottom sheet mobile mendukung swipe-down close dengan threshold + snap-back

### Form Lapor (`/lapor`)

- Title 20px/700, section spacing 24px
- Severity selected `#E5484D`, bg `#F1F5F9`
- Submit `#E5484D`, disabled 0.45, active scale(0.97)
- Semua px units, warna sesuai spec

### Detail Laporan (`/issues/[id]`)

- Layout mobile-first dengan urutan informasi:
  1. hero media utama (full-width mobile, rounded desktop)
  2. badges severity/status/verification
  3. metrik ringkas (laporan/foto/korban/reaksi/visibility)
  4. info utama + info tambahan + catatan publik (jika ada)
  5. galeri media + aktivitas terbaru + CTA share/action
- Shareability built-in:
  - tombol share
  - link cepat WhatsApp/Telegram/Twitter(X)/Facebook
  - metadata SEO/OG/Twitter canonical disiapkan dari SSR data route
- Fallback states wajib:
  - not found
  - error
  - loading retry
  - fallback media rusak
  - empty gallery
- Visual tetap mengikuti token:
  - severity colors (kuning-oranye-merah)
  - status/verification badge color-coded
  - card white + border `#E2E8F0`, radius 16px
  - CTA primary `#E5484D`, min-height 48px

---

## Pola Perubahan Umum

| Sebelum              | Sesudah     | Alasan                          |
| -------------------- | ----------- | ------------------------------- |
| `#e53e3e`            | `#E5484D`   | JEDUG Red brand konsisten       |
| `#4a5568`, `#718096` | `#64748B`   | Neutral text secondary          |
| `#a0aec0`            | `#94A3B8`   | Muted text                      |
| `#2d3748`            | `#0F172A`   | Text primary                    |
| `#f7fafc`            | `#F8FAFC`   | Background                      |
| `#edf2f7`            | `#F1F5F9`   | Surface                         |
| `#e2e8f0`            | `#E2E8F0`   | Border                          |
| `rem` units          | `px` units  | Konsistensi, predictable sizing |
| Emoji di CTA         | Text only   | Lebih profesional               |
| Status 1 warna       | Color-coded | Informatif                      |

---

## Apa yang Belum Diubah

- **Flow map publik utama (`/issues`)** — tetap map-first, tidak dirombak.
- **Flow submit report (`/lapor`)** — tidak berubah.
- **Tailwind/CSS framework** — tetap inline scoped styles (tanpa migrasi framework).

## Read This Next

- `design-docs/design-system.md`
- `design-docs/component-spec.md`
- `docs/DESIGN_INDEX.md`
