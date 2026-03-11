# Changelog for Agents

Tujuan file ini: ringkasan perubahan teknis lintas area yang wajib diketahui agent berikutnya.

## Aturan Update Wajib

Setiap agent yang mengubah area signifikan project WAJIB:

1. update dokumen teknis terkait di `docs/`
2. update file ini (`docs/CHANGELOG_FOR_AGENTS.md`)
3. update `AGENTS.md` jika onboarding/flow kerja berubah
4. update `design-docs/` jika menyentuh design system, behavior komponen, atau page-level UX
5. catat keputusan arsitektur/produk baru di `docs/DECISIONS.md`

Area yang selalu wajib update docs bila berubah:

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

## 2026-03-10 - Map Marker Clustering + Location Label UX (`/lapor`)

- Scope:
  - mengganti render marker map publik ke `GeoJSON source + layer` MapLibre dengan clustering.
  - menambahkan interaksi cluster (count, click-to-zoom) dan menjaga marker individual tetap clickable untuk bottom sheet.
  - menambahkan endpoint backend ringan `GET /api/v1/location/label` untuk resolve label wilayah dari `regions`.
  - menambahkan UX label lokasi manusiawi di `/lapor` dengan cache/guard request, tanpa memblok submit jika lookup gagal.
- Dampak area:
  - `frontend/src/lib/components/IssueMap.svelte`
  - `frontend/src/routes/lapor/+page.svelte`
  - `frontend/src/lib/api/location.ts`
  - `frontend/src/lib/api/types.ts`
  - `backend/internal/http/router.go`
  - `backend/internal/http/handlers/location.go`
  - `backend/internal/service/location_service.go`
  - `backend/internal/repository/location_repository.go`
- File docs yang diupdate:
  - `docs/BACKEND.md`
  - `docs/FRONTEND.md`
  - `docs/MAP_AND_LOCATION.md`
  - `design-docs/component-spec.md`
  - `design-docs/guide.md`
  - `docs/CHANGELOG_FOR_AGENTS.md`
- Mismatch baru (jika ada):
  - tidak ada mismatch arsitektur baru; endpoint label lokasi bersifat additive dan non-blocking terhadap flow submit.

## 2026-03-10 - Duplicate Detection / Smart Issue Merge di Submit Report

- Scope:
  - memperkuat flow `POST /api/v1/reports` agar submission baru merge ke issue aktif terdekat, bukan selalu membuat issue baru.
  - menambah duplicate radius configurable (`DUPLICATE_RADIUS_M`, default 30 meter).
  - menambahkan tie-break kandidat (distance -> status/verification -> last_seen -> severity) + logging audit behavior merge/new issue.
  - mengubah agregasi merge issue untuk `casualty_count` menjadi `GREATEST(existing, incoming)` agar tidak overcount duplikasi.
- Dampak area:
  - `backend/internal/repository/report_repository.go`
  - `backend/internal/service/report_service.go`
  - `backend/internal/http/router.go`
  - `backend/internal/config/config.go`
  - `backend/internal/repository/report_repository_test.go`
  - `backend/internal/http/handlers/report_handler_test.go`
  - `backend/.env.example`
- File docs yang diupdate:
  - `docs/PROJECT_OVERVIEW.md`
  - `docs/ARCHITECTURE.md`
  - `docs/BACKEND.md`
  - `docs/SCHEMA.md`
  - `docs/DEPLOYMENT.md`
  - `docs/DECISIONS.md`
  - `docs/CHANGELOG_FOR_AGENTS.md`
- Mismatch baru (jika ada):
  - test integration ke Postgres/PostGIS belum ada; skenario geo query multi-kandidat masih perlu verifikasi manual/staging DB.

## 2026-03-10 - Baseline Dokumentasi Terpadu

- Menambahkan pusat navigasi agent di `AGENTS.md`.
- Menyatukan dokumentasi teknis utama ke folder `docs/`.
- Menetapkan `design-docs/` sebagai source of truth desain dan menghubungkannya lewat `docs/DESIGN_INDEX.md`.
- Mendokumentasikan flow deploy aktual berbasis GitHub Actions + PM2.
- Mendokumentasikan storage local vs R2 dan strategi fallback media legacy.
- Mendokumentasikan flow moderation admin + community flag auto-hide.
- Menambahkan penjelasan schema lintas tabel berikut konsep `issue` vs `issue_submission`.

## Mismatch yang Teridentifikasi (Perlu Tindak Lanjut)

- SQL schema source masih berada di luar repo (`/Users/ryanprayoga/Downloads/jedug_schema_v2.sql`).
- Konfigurasi PM2/Nginx runtime belum versioned di repo.
- auth admin runtime masih env + in-memory session, belum memakai tabel user/session schema.
- indikasi formatting typo pada SQL `submission_media` (`widthINT/heightINT`) perlu verifikasi manual terhadap DB nyata.

## 2026-03-10 - Issue Detail Page Production-Ready

- Scope:
  - merapikan ulang halaman publik `/issues/[id]` menjadi halaman detail issue yang mobile-first, shareable, dan SEO friendly.
  - memecah UI detail issue menjadi komponen `IssueHeader`, `IssueStats`, `IssueGallery`, dan `ShareActions`.
  - memperketat endpoint publik detail issue agar konsisten dengan visibilitas publik map/list.
- Dampak area:
  - `frontend/src/routes/issues/[id]/+page.ts`
  - `frontend/src/routes/issues/[id]/+page.svelte`
  - `frontend/src/routes/+layout.svelte`
  - `frontend/src/lib/components/IssueHeader.svelte`
  - `frontend/src/lib/components/IssueStats.svelte`
  - `frontend/src/lib/components/IssueGallery.svelte`
  - `frontend/src/lib/components/ShareActions.svelte`
  - `frontend/src/lib/utils/issue-detail.ts`
  - `backend/internal/repository/issue_repository.go`
- File docs yang diupdate:
  - `docs/BACKEND.md`
  - `docs/FRONTEND.md`
  - `design-docs/component-spec.md`
  - `design-docs/guide.md`
- Mismatch baru (jika ada):
  - endpoint publik detail issue menampilkan maksimal 20 foto publik terbaru demi payload yang tetap ringan; UI menjelaskan bila total foto lebih besar dari subset yang ditampilkan.

## 2026-03-10 - Production-Ready Public Issue Detail + Shareability

- Scope:
  - memperkuat halaman publik `/issues/[id]` agar mobile-first, informatif, dan siap dibagikan ke sosial media.
  - menambahkan metadata SEO/OG/Twitter + canonical berbasis SSR route data.
  - menambahkan `region_name` pada response issue publik/admin (derived field via join `regions`).
- Dampak area:
  - frontend route issue detail (`+page.ts` dan `+page.svelte`)
  - helper UI/SEO share (`src/lib/utils/issue-detail.ts`)
  - backend repository/domain issue dan admin issue (query join region)
  - static asset fallback OG image (`frontend/static/og/issue-fallback.svg`)
- File docs yang diupdate:
  - `docs/FRONTEND.md`
  - `docs/BACKEND.md`
  - `docs/SCHEMA.md`
  - `design-docs/guide.md`
- Mismatch baru (jika ada):
  - tidak ada mismatch baru yang teridentifikasi dari perubahan ini; kontrak lama tetap dipertahankan dan hanya ditambah field turunan.

## 2026-03-10 - UX Bugfix Map Mobile + CTA/Homepage Polish

- Scope:
  - memperbaiki interaksi `IssueBottomSheet` agar bisa swipe/drag down untuk close di mobile.
  - menstabilkan transisi list ↔ map dengan guard state agar tidak muncul false-empty/flicker.
  - mengganti empty state map dari popup tengah menjadi info badge top-left agar tidak menutupi peta.
  - memoles visual CTA utama (`Laporkan Jalan Rusak`) di map/home dan komponen tombol utama.
  - memoles homepage agar hierarki visual lebih matang tanpa mengubah flow bisnis.
- Dampak area:
  - `frontend/src/routes/issues/+page.svelte`
  - `frontend/src/lib/components/IssueBottomSheet.svelte`
  - `frontend/src/lib/utils/bbox.ts`
  - `frontend/src/lib/components/PrimaryButton.svelte`
  - `frontend/src/routes/+page.svelte`
- File docs yang diupdate:
  - `docs/FRONTEND.md`
  - `docs/MAP_AND_LOCATION.md`
  - `design-docs/component-spec.md`
  - `design-docs/guide.md`
- Mismatch baru (jika ada):
  - tidak ada mismatch arsitektur baru; perubahan terbatas pada UX state dan visual polish frontend.

## 2026-03-10 - Issue Detail Public Contract Tightening

- Scope:
  - menambah field additive aman pada `GET /api/v1/issues/:id` untuk halaman detail publik production-ready.
  - mempertegas urutan informasi `/issues/[id]` agar lebih siap dibagikan dan lebih aman untuk publik.
- Dampak area:
  - `backend/internal/domain/issue.go`
  - `backend/internal/domain/submission.go`
  - `backend/internal/repository/issue_repository.go`
  - `backend/internal/http/handlers/issue.go`
  - `frontend/src/lib/api/types.ts`
  - `frontend/src/lib/utils/issue-detail.ts`
  - `frontend/src/lib/components/IssueHeader.svelte`
  - `frontend/src/lib/components/IssueStats.svelte`
  - `frontend/src/routes/issues/[id]/+page.svelte`
- File docs yang diupdate:
  - `docs/BACKEND.md`
  - `docs/FRONTEND.md`
  - `design-docs/component-spec.md`
  - `design-docs/guide.md`
  - `docs/CHANGELOG_FOR_AGENTS.md`
- Mismatch baru (jika ada):
  - tidak ada perubahan breaking; note mentah tetap ada di response untuk kompatibilitas, tetapi UI publik dipindahkan ke `public_note` yang sudah diringkas.

## 2026-03-10 - Dynamic OG Image Generator untuk Issue

- Scope:
  - menambahkan endpoint server-side OG image dinamis untuk issue publik: `/api/og/issues/[id]`.
  - memperbarui metadata SEO issue detail agar `og:image` dan `twitter:image` selalu memakai endpoint OG dinamis.
- Dampak area:
  - `frontend/src/routes/api/og/issues/[id]/+server.ts`
  - `frontend/src/routes/issues/[id]/+page.ts`
  - `frontend/src/routes/issues/[id]/+page.svelte`
  - `frontend/src/lib/utils/issue-detail.ts`
  - `frontend/package.json`
  - `frontend/package-lock.json`
- File docs yang diupdate:
  - `docs/FRONTEND.md`
  - `design-docs/guide.md`
  - `docs/CHANGELOG_FOR_AGENTS.md`
- Mismatch baru (jika ada):
  - asumsi runtime frontend production berbasis adapter-node (bukan edge runtime) agar rendering `@vercel/og` berjalan konsisten di VPS + PM2.

## Template Entri Berikutnya

Gunakan format ini untuk update berikutnya:

`## YYYY-MM-DD - Judul Perubahan`

- Scope:
- Dampak area:
- File docs yang diupdate:
- Mismatch baru (jika ada):

## Read This Next

- `AGENTS.md`
- `docs/DECISIONS.md`
