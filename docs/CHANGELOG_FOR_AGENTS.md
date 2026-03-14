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

## 2026-03-14 - Submit Report 500 Fix: issue_events Migration + Non-Fatal Event Inserts

- Akar masalah:
  - Tabel `issue_events` belum pernah dibuat di database (migrations/ kosong).
  - `createIssueEvent()` dipanggil *di dalam* transaction utama submit report.
  - Query INSERT ke tabel yang tidak ada melempar DB error, menyebabkan seluruh TX rollback.
  - Handler menangkap error ini sebagai generic 500 "failed to submit report".
- Perbaikan:
  1. Membuat file migrasi `backend/migrations/202603140001_create_issue_events.sql` — wajib dijalankan di production dengan `psql $DATABASE_URL -f migrations/202603140001_create_issue_events.sql`.
  2. Memindahkan seluruh `createIssueEvent` calls keluar dari TX utama ke method baru `insertTimelineEvents()` yang berjalan setelah `tx.Commit()` menggunakan `r.db` (pool langsung).
  3. Event insert kini non-fatal: error hanya di-log (`[REPORT] timeline_event_insert_error`), tidak pernah membatalkan atau mengembalikan error ke caller.
  4. Menghapus fungsi `createIssueEvent(ctx, tx, ...)` yang kini menjadi dead code.
- Dampak area:
  - `backend/migrations/202603140001_create_issue_events.sql` (baru — WAJIB RUN DI PROD)
  - `backend/internal/repository/report_repository.go`
- File docs yang diupdate:
  - `docs/CHANGELOG_FOR_AGENTS.md`
- Action yang wajib dilakukan setelah deploy:
  - Jalankan migration di production database sebelum atau sesaat setelah deploy binary baru.
  - Monitor log `[REPORT] timeline_event_insert_error` untuk memastikan events mulai tercatat setelah migration dijalankan.

## 2026-03-14 - Submit Report Error Handling Overhaul

- Scope:
  - mengaudit end-to-end pipeline submit laporan dari bootstrap → compress → presign → upload → submit → success/error.
  - menemukan root cause utama: catch block di `handleSubmit` menggunakan `e.message` mentah untuk semua error yang tidak dikelola secara eksplisit, sehingga pesan seperti "failed to submit report" (HTTP 500), "device is banned" (HTTP 403), dan "Failed to fetch" (network error) langsung tampil ke user.
  - menambahkan fungsi `mapSubmitError(e: unknown): string` untuk memetakan semua error (ApiError 4xx/5xx, TypeError network, Error pipeline) ke pesan Indonesia yang kontekstual dan ramah pengguna.
  - mengganti catch block di `handleSubmit` dari rantai if/else per status ke single `mapSubmitError` call dengan scroll-to-error otomatis.
  - menambahkan `console.error` di catch untuk debugging submit failure.
  - menambahkan `log.Printf` di backend handler sebelum return HTTP 500 agar error aktual ter-log di server.
- Error mapping baru:
  - `TypeError` (network/fetch failure) → "Koneksi sedang bermasalah. Periksa internet lalu coba lagi."
  - `ApiError 429` → pesan backend (sudah dalam bahasa Indonesia)
  - `ApiError 403` → "Akun tidak diizinkan mengirim laporan saat ini."
  - `ApiError 401` → "Inisialisasi pelaporan belum selesai. Muat ulang halaman lalu coba lagi."
  - `ApiError 400` → "Data laporan belum valid. Periksa kembali isian form lalu coba lagi."
  - `ApiError 5xx` → "Laporan belum bisa dikirim saat ini. Coba beberapa saat lagi."
  - `Error 'Canvas context not available'` → "Gagal memproses foto. Coba pilih ulang foto lalu kirim lagi."
  - Error Indonesia yang dilempar pipeline (compress/upload) → pass-through as-is
- Dampak area:
  - `frontend/src/routes/lapor/+page.svelte`
  - `backend/internal/http/handlers/report_handler.go`
- File docs yang diupdate:
  - `docs/CHANGELOG_FOR_AGENTS.md`
- Mismatch baru (jika ada):
  - tidak ada perubahan kontrak API; semua perubahan di sisi error presentation layer.

## 2026-03-14 - Core UX Bugfixes (Lapor Bootstrap, Blue Dot Map, Active Navbar)

- Scope:
  - memperbaiki race condition bootstrap device anonim yang menyebabkan submit report gagal dengan pesan `device not found; bootstrap first`.
  - menambahkan guard bootstrap eksplisit pada `/lapor` + retry ringan + retry submit otomatis sekali saat token bootstrap mismatch.
  - memperbaiki UX map agar geolocate dipicu otomatis sekali pada first load (blue dot langsung tampil jika izin lokasi tersedia) tanpa recenter berulang.
  - memperbaiki active state navbar agar langsung sinkron dengan route pada initial render/refresh, bukan click-only.
- Dampak area:
  - `frontend/src/lib/utils/device-init.ts`
  - `frontend/src/routes/+layout.svelte`
  - `frontend/src/routes/lapor/+page.svelte`
  - `frontend/src/lib/components/IssueMap.svelte`
  - `frontend/src/lib/components/AppHeader.svelte`
- File docs yang diupdate:
  - `docs/FRONTEND.md`
  - `docs/CHANGELOG_FOR_AGENTS.md`
- Mismatch baru (jika ada):
  - tidak ada mismatch kontrak API; perbaikan fokus ke sequencing/bootstrap guard, geolocate initialization, dan route-driven UI state.

## 2026-03-14 - Issue Timeline / Riwayat Perkembangan Issue

- Scope:
  - menambahkan tabel `issue_events` sebagai audit timeline publik per issue.
  - menambahkan endpoint publik `GET /api/v1/issues/:id/timeline` dengan pagination (`limit`, `offset`) dan ordering terbaru di atas.
  - menambahkan event logging otomatis untuk `issue_created`, `photo_added`, `severity_changed`, `casualty_reported`, `status_updated`.
  - menambahkan section UI `Riwayat Laporan` pada detail issue publik (`/issues/[id]`) dengan timeline vertikal mobile-first + load more saat event > 100.
- Dampak area:
  - `backend/migrations/202603140001_create_issue_events.sql`
  - `backend/internal/repository/report_repository.go`
  - `backend/internal/repository/admin_repository.go`
  - `backend/internal/repository/issue_repository.go`
  - `backend/internal/service/issue_service.go`
  - `backend/internal/http/handlers/issue.go`
  - `backend/internal/http/router.go`
  - `backend/internal/domain/issue.go`
  - `frontend/src/lib/api/types.ts`
  - `frontend/src/lib/api/issues.ts`
  - `frontend/src/routes/issues/[id]/+page.svelte`
- File docs yang diupdate:
  - `docs/SCHEMA.md`
  - `docs/BACKEND.md`
  - `docs/FRONTEND.md`
  - `design-docs/guide.md`
  - `docs/CHANGELOG_FOR_AGENTS.md`
- Mismatch baru (jika ada):
  - baseline schema SQL utama project masih direferensikan dari file eksternal, namun migration additive untuk timeline (`issue_events`) kini sudah versioned di repo.

## 2026-03-12 - Public Map Heatmap / Severity Visualization

- Scope:
  - menambahkan mode visual `Heatmap` pada halaman publik `/issues` tanpa mengubah endpoint backend.
  - mempertahankan mode `Marker` + clustering yang sudah live sebagai mode default dan stable.
  - menambahkan formula weight ringan berbasis `severity_current`, `casualty_count`, `submission_count`, dan penurunan bobot untuk status `fixed/archived`.
  - mengelola marker source dan heatmap source via `setData` + toggle `visibility` agar perpindahan mode stabil dan tidak add/remove layer berulang saat user klik toggle.
  - menambahkan fallback non-blocking: bila heatmap gagal dimuat, UI kembali ke marker mode dan peta tetap usable.
- Dampak area:
  - `frontend/src/lib/components/IssueMap.svelte`
  - `frontend/src/routes/issues/+page.svelte`
  - `frontend/src/lib/utils/issue-heatmap.ts`
- File docs yang diupdate:
  - `docs/FRONTEND.md`
  - `docs/MAP_AND_LOCATION.md`
  - `design-docs/design-system.md`
  - `design-docs/component-spec.md`
  - `design-docs/guide.md`
  - `docs/CHANGELOG_FOR_AGENTS.md`
- Mismatch baru (jika ada):
  - tidak ada mismatch kontrak API; heatmap seluruhnya dibangun dari field issue publik yang sudah tersedia.

## 2026-03-12 - Bugfix Location Label `/lapor` (internal region miss -> reverse fallback)

- Scope:
  - memperbaiki pipeline `GET /api/v1/location/label` agar tidak berhenti di lookup internal `regions` saja.
  - menambahkan fallback reverse geocoding pada service label lokasi ketika region internal tidak ditemukan.
  - menambah debug log end-to-end untuk audit koordinat masuk, hasil query geospatial, trigger reverse, dan response akhir.
  - menambah fallback provider reverse geocode sekunder agar label tetap tersedia saat provider utama error.
- Dampak area:
  - `backend/internal/http/handlers/location.go`
  - `backend/internal/repository/location_repository.go`
  - `backend/internal/service/location_service.go`
  - `backend/internal/service/location_service_test.go`
  - `backend/internal/service/reverse_geocoder.go`
  - `backend/internal/http/router.go`
  - `frontend/src/routes/lapor/+page.svelte`
- File docs yang diupdate:
  - `docs/BACKEND.md`
  - `docs/FRONTEND.md`
  - `docs/CHANGELOG_FOR_AGENTS.md`
- Mismatch baru (jika ada):
  - jika tabel `regions` kosong, source label lokasi akan bergantung pada reverse geocoding; jika semua provider reverse gagal, endpoint tetap return sukses dengan `source=unresolved` dan label `null`.

## 2026-03-12 - Location Normalization saat Submit Report + Entry Point `/stats`

- Scope:
  - menambahkan normalisasi lokasi saat `POST /api/v1/reports` untuk memperkaya `road_name` dan `region` issue publik.
  - menambahkan reverse geocoding ringan (timeout + cache in-memory) yang hanya dipanggil saat submit report.
  - memastikan merge duplicate tidak overwrite data lokasi valid, hanya melengkapi field kosong.
  - menambahkan helper command backfill issue lama yang masih kosong lokasi.
  - menambahkan entry point UI menuju `/stats` di header dan homepage.
- Dampak area:
  - `backend/internal/config/config.go`
  - `backend/internal/service/reverse_geocoder.go`
  - `backend/internal/service/report_location_normalizer.go`
  - `backend/internal/service/report_service.go`
  - `backend/internal/repository/report_repository.go`
  - `backend/internal/http/router.go`
  - `backend/cmd/backfill_issue_location/main.go`
  - `backend/.env.example`
  - `frontend/src/lib/components/AppHeader.svelte`
  - `frontend/src/routes/+page.svelte`
  - `frontend/src/routes/lapor/+page.svelte`
- File docs yang diupdate:
  - `docs/BACKEND.md`
  - `docs/FRONTEND.md`
  - `docs/DEPLOYMENT.md`
  - `design-docs/component-spec.md`
  - `design-docs/guide.md`
  - `docs/CHANGELOG_FOR_AGENTS.md`
- Mismatch baru (jika ada):
  - schema issue saat ini tidak punya kolom `region_name`/`city_name` fisik; nilai wilayah publik tetap diturunkan dari `region_id` + tabel `regions`.

## 2026-03-12 - Public Stats Dashboard (`/stats`) + API Aggregation Endpoint

- Scope:
  - menambahkan endpoint publik baru `GET /api/v1/stats` untuk statistik agregasi issue.
  - menambahkan in-memory cache ringan di backend (TTL 45 detik) agar query agregasi tidak menghantam DB setiap request.
  - menambahkan halaman publik baru `/stats` (mobile-first) berisi global stats, status breakdown, time stats, region leaderboard, dan top issue.
- Dampak area:
  - `backend/internal/http/router.go`
  - `backend/internal/http/handlers/stats.go`
  - `backend/internal/service/stats_service.go`
  - `backend/internal/service/stats_service_test.go`
  - `backend/internal/repository/stats_repository.go`
  - `backend/internal/domain/stats.go`
  - `frontend/src/lib/api/stats.ts`
  - `frontend/src/lib/api/types.ts`
  - `frontend/src/routes/stats/+page.svelte`
- File docs yang diupdate:
  - `docs/BACKEND.md`
  - `docs/FRONTEND.md`
  - `design-docs/guide.md`
  - `docs/CHANGELOG_FOR_AGENTS.md`
- Mismatch baru (jika ada):
  - backend aktif project ini tetap Go + Fiber; deskripsi eksternal yang menyebut backend Node tidak merefleksikan implementasi repo saat ini.

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
