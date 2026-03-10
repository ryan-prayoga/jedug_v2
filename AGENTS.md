# JEDUG Agent Guide (Single Source of Truth)

Dokumen ini adalah pintu masuk utama untuk agent mana pun yang bekerja di JEDUG.

## 1) Apa Itu JEDUG

JEDUG adalah aplikasi civic-tech untuk pelaporan jalan rusak berbasis peta, dengan alur pelaporan cepat (anonim), moderasi admin, serta publikasi issue secara terbuka.

## 2) Status Project Saat Ini (2026-03-10)

- Stack utama: Go + Fiber + PostgreSQL/PostGIS (backend), SvelteKit 2 + Svelte 5 + MapLibre (frontend).
- Produk sudah berjalan end-to-end untuk:
  - bootstrap device anonim
  - consent
  - upload media (local/R2)
  - submit report
  - list/detail issue publik
  - admin moderation (hide/unhide/fix/reject/ban)
- Dokumentasi desain tersedia di `design-docs/`.
- Dokumentasi teknis terpadu ada di `docs/`.

## 3) Struktur Repo Tingkat Tinggi

- `backend/`: API, service domain, repository SQL, middleware, storage local/R2.
- `frontend/`: aplikasi publik + admin (SvelteKit), peta, pelaporan, moderasi UI.
- `design-docs/`: source of truth desain UI/UX (token, komponen, page guidance).
- `docs/`: source of truth teknis/arsitektur/operasional/schema/keputusan.
- `.github/workflows/deploy.yml`: deploy CI/CD ke VPS via SSH.

## 4) Dokumen Wajib Dibaca Sebelum Coding

1. `AGENTS.md` (file ini)
2. `docs/PROJECT_OVERVIEW.md`
3. `docs/ARCHITECTURE.md`
4. `docs/SCHEMA.md`
5. `docs/BACKEND.md` atau `docs/FRONTEND.md` (sesuai task)
6. `docs/DESIGN_INDEX.md` + file terkait di `design-docs/`

## 5) Read Order yang Disarankan

1. Mulai dari `AGENTS.md`
2. Pahami konteks bisnis di `docs/PROJECT_OVERVIEW.md`
3. Pahami aliran sistem di `docs/ARCHITECTURE.md`
4. Pahami model data di `docs/SCHEMA.md`
5. Masuk ke area implementasi:
   - backend: `docs/BACKEND.md`
   - frontend: `docs/FRONTEND.md`
   - deploy/storage/moderasi/map: file tematik di `docs/`
6. Sinkronkan keputusan UI/UX lewat `docs/DESIGN_INDEX.md` dan `design-docs/*`

## 6) Aturan Kerja Agent

- Gunakan implementasi aktual sebagai acuan, bukan arsitektur ideal.
- Jika ada gap, tulis eksplisit:
  - `Current implementation`
  - `Intended direction`
  - `Known mismatch`
- Jangan duplikasi dokumentasi desain ke `docs/`; lakukan referensi silang.
- Jangan ubah kontrak response API tanpa mengecek dampak ke:
  - map markers
  - issue detail publik
  - admin issue detail
- Hindari perubahan diam-diam pada flow upload/media/moderation tanpa update docs.

## 7) Checklist Sebelum Mengubah Kode

- Sudah baca dokumen area terkait (`docs/*`) dan dokumen desain terkait (`design-docs/*`).
- Sudah petakan dampak perubahan ke endpoint, schema, dan response shape.
- Sudah identifikasi area sensitif:
  - `issue` response shape
  - upload/presign/public URL
  - map bbox + marker selection + bottom sheet
  - moderation visibility (is_hidden/status)
- Sudah tentukan dokumen mana yang wajib di-update setelah perubahan.

## 8) Checklist Setelah Mengubah Kode

- Verifikasi flow utama area yang diubah (minimal manual smoke test).
- Perbarui dokumentasi teknis terkait di `docs/`.
- Tambahkan entri ringkas di `docs/CHANGELOG_FOR_AGENTS.md`.
- Jika onboarding/flow kerja agent berubah, update `AGENTS.md`.
- Jika menyentuh desain, update `design-docs/*` yang relevan.
- Jika ada keputusan arsitektur/produk baru, catat di `docs/DECISIONS.md`.

## 9) Aturan Update Dokumentasi (WAJIB)

Setiap perubahan signifikan WAJIB meng-update dokumen terkait untuk area berikut:

- route/endpoints
- schema database
- env variables
- deployment flow
- storage/media flow
- moderation logic
- map behavior
- anti-spam/trust logic
- struktur repo
- UI system/component rules
- `design-docs/*` jika mengubah visual style, component behavior, atau page-level UX

Aturan tambahan untuk agent berikutnya:

1. update dokumen teknis terkait di `docs/`
2. update `docs/CHANGELOG_FOR_AGENTS.md`
3. update `AGENTS.md` bila perubahan mempengaruhi onboarding/flow kerja
4. update `design-docs/` bila perubahan menyentuh design system, behavior komponen, atau UX level halaman
5. bila ada keputusan arsitektur baru, catat di `docs/DECISIONS.md`

## 10) Context for Future Agents

### Area Relatif Stabil

- Struktur backend `handler -> service -> repository`.
- Konsep identity anonim berbasis `device` + `anon_token_hash`.
- Struktur route publik utama (`/issues`, `/reports`, `/uploads`, `/device`).

### Area Sering Berubah

- UI/UX publik (map/list/sheet/forms).
- Deployment dan operasi server (PM2/Nginx/env VPS).
- Integrasi storage (local <-> R2) dan public URL strategy.

### Area Sensitif / Gampang Rusak

- Shape response `Issue` berdampak langsung ke map + detail + admin.
- Upload flow berdampak ke submit report + rendering media.
- Moderation logic berdampak ke visibility publik (`is_hidden`, `status`).
- BBox/map marker behavior sensitif pada field location/status/severity.
- Konsistensi `design-docs/` dengan komponen aktual frontend.

### Area Paling Sering Memicu Bug

- perbedaan status issue vs visibility moderasi
- fallback upload saat mode R2
- perbedaan sesi admin (in-memory) vs ekspektasi persistence
- mismatch schema SQL vs query repository

### Source of Truth

- Desain/UI: `docs/DESIGN_INDEX.md` + `design-docs/design-system.md`, `design-docs/component-spec.md`, `design-docs/guide.md`
- Arsitektur/teknis: `docs/ARCHITECTURE.md`, `docs/BACKEND.md`, `docs/FRONTEND.md`, `docs/SCHEMA.md`, `docs/DEPLOYMENT.md`

## 11) When In Doubt

Jika ragu:

1. berhenti mengasumsikan
2. cek implementasi kode aktual
3. cek dokumen source of truth yang relevan
4. dokumentasikan gap sebagai `Known mismatch`
5. pilih perubahan paling aman untuk backward compatibility

## Read This Next

- `docs/PROJECT_OVERVIEW.md`
- `docs/ARCHITECTURE.md`
- `docs/SCHEMA.md`
- `docs/DESIGN_INDEX.md`
