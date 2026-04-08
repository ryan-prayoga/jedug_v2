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

## 2026-04-08 - MapLibre Runtime Dikeluarkan dari Chunk Vite

- Scope:
  - menghilangkan warning chunk besar pada route `/issues` tanpa mengubah kontrak API atau flow peta publik.

### Frontend

1. `IssueMap.svelte` tidak lagi memuat runtime `maplibre-gl` lewat `import()` bundler.
2. Runtime MapLibre sekarang dimuat sebagai asset script eksternal (`?url`) saat route peta benar-benar dibuka, dan stylesheet-nya juga diinject on-demand.
3. Route `/issues` tetap menampilkan loading overlay lokal sambil runtime peta diunduh dan diinisialisasi.

### Dampak

4. Build frontend tidak lagi menghasilkan warning `Some chunks are larger than 500 kB after minification` dari MapLibre.
5. Perubahan ini mengecilkan graph chunk JS aplikasi, tetapi tidak mengecilkan bobot runtime MapLibre itu sendiri secara material; biaya download peta tetap dominan saat user benar-benar membuka `/issues`.

### Docs

6. Diperbarui:
   - `docs/FRONTEND.md`
   - `docs/CHANGELOG_FOR_AGENTS.md`

## 2026-04-08 - Map Runtime Lazy-Load + Metadata Fix `/lapor`

- Scope:
  - mempercepat shell `/issues` saat route peta dibuka tanpa mengubah kontrak API, sekaligus menutup bug metadata client-side pada `/lapor`.

### Frontend

1. `IssueMap.svelte` tidak lagi mengimpor `maplibre-gl` dan CSS-nya secara eager di entry komponen.
2. Runtime MapLibre kini dilazy-load saat komponen mount, sehingga wrapper route `/issues` bisa tampil dulu sambil library peta diunduh.
3. Ditambahkan overlay loading ringan di level komponen peta agar state sebelum runtime MapLibre siap tetap jelas bagi user.
4. `routes/lapor/+page.svelte` sekarang punya `svelte:head` sendiri (`title` + `meta description`), jadi navigasi dari `/stats` atau route lain tidak lagi meninggalkan title lama di browser.

### Dampak

5. Perubahan ini memperbaiki jalur loading `/issues`, tetapi tidak menghilangkan fakta bahwa byte utama MapLibre tetap besar; warning build hanya akan hilang penuh jika library dieksternalisasi atau stack peta diganti.
6. Smoke test browser manual di preview lokal (dengan mock API) berhasil melewati `/`, notification center, `/issues`, `/issues/[id]`, `/stats`, dan `/lapor`; temuan runtime yang tersisa hanya `favicon.ico 404`.

### Docs

7. Diperbarui:
   - `docs/FRONTEND.md`
   - `docs/CHANGELOG_FOR_AGENTS.md`

## 2026-04-08 - Notification Center Public Shell Dilazy-Load

- Scope:
  - menurunkan biaya bundle shared publik tanpa mengubah perilaku utama notification center.

### Frontend

1. `routes/+layout.svelte` tidak lagi menginisialisasi browser push dan notification preferences pada first paint; layout kini hanya bootstrap device + in-app notifications.
2. `AppHeader.svelte` tidak lagi mengimpor `BrowserPushCard`, `NotificationPreferencesPanel`, dan `NearbyAlertsPanel` secara statis.
3. Tiga panel notification center tersebut sekarang di-load lazy saat user benar-benar membuka panel lonceng.
4. `BrowserPushCard.svelte` dan `NotificationPreferencesPanel.svelte` kini self-initializing agar tetap aman dipakai di route lain tanpa bergantung pada root layout.

### Dampak

5. Shared chunk publik berkurang karena kode browser push/preferences tidak lagi ikut first paint semua route non-admin.
6. Warning bundle terbesar tetap berasal dari chunk MapLibre di route `/issues`, jadi optimasi ini memang menyasar biaya shared shell, bukan chunk peta.

### Docs

7. Diperbarui:
   - `docs/FRONTEND.md`
   - `docs/CHANGELOG_FOR_AGENTS.md`

## 2026-04-08 - Public Issue Visibility dan Enum Status Diselaraskan ke Schema

- Scope:
  - menutup kebocoran submission `spam` pada jalur publik sekaligus membersihkan drift enum issue/verification antara backend, stats, dan presenter frontend.

### Backend

1. `IssueRepository` sekarang hanya memakai submission non-`rejected`/`spam` untuk:
   - galeri media publik
   - `recent_submissions`
   - fallback label wilayah dari latest submission
2. `ReportRepository` duplicate detection kini memakai issue aktif kanonik `status = 'open'`.
3. Ranking duplicate diperbarui agar memahami verification kanonik `admin_verified/community_verified/unverified`, dengan fallback legacy tetap defensif.
4. `StatsRepository` tidak lagi menghitung status historis `verified/in_progress` sebagai issue open dan juga mengabaikan submission `spam/rejected` saat menurunkan label wilayah.

### Frontend / Design

5. Presenter status/verification publik sekarang eksplisit mendukung enum schema:
   - status issue `open/fixed/archived/rejected/merged`
   - verification `unverified/community_verified/admin_verified`
6. `IssueCard` dan `IssueBottomSheet` ikut diselaraskan agar badge publik tidak jatuh ke raw string saat menerima enum kanonik.
7. Design docs status/verification diperbarui agar sinkron dengan implementasi.

### Docs

8. Diperbarui:
   - `docs/BACKEND.md`
   - `docs/FRONTEND.md`
   - `docs/MAP_AND_LOCATION.md`
   - `docs/ARCHITECTURE.md`
   - `docs/SCHEMA.md`
   - `docs/DECISIONS.md`
   - `backend/schema/README.md`
   - `design-docs/design-system.md`
   - `design-docs/guide.md`
   - `docs/CHANGELOG_FOR_AGENTS.md`

## 2026-03-20 - Tailwind CSS + Iconify Frontend Polish

- Scope:
  - merapikan frontend publik + admin ke design system yang konsisten tanpa mengubah flow fitur inti yang sudah jalan.

### Frontend / UI Foundation

1. Frontend sekarang memakai Tailwind CSS v4 via plugin Vite (`@tailwindcss/vite`) dan source of truth visual global ada di `frontend/src/app.css`.
2. Ditambahkan adapter icon tunggal `frontend/src/lib/icons.ts` dengan family Iconify Solar `line-duotone`.
3. Font UI digeser ke `Plus Jakarta Sans` lewat `frontend/src/app.html`.
4. Primitive visual reusable baru/distandarkan:
   - shell/layout: `app-shell`, `app-main`, `app-main-wide`, `app-main-full`
   - surface: `jedug-card`, `jedug-card-soft`, `jedug-panel`, `admin-card`
   - form/button/badge/state classes di `app.css`

### Public Pages

5. Header publik, landing, form lapor, notification center, map bottom sheet, issue cards, gallery, share actions, dan state components direfresh ke visual yang lebih modern dan konsisten.
6. Route `/issues` sekarang punya shell Tailwind penuh:
   - hero summary card
   - segmented toggle marker/heatmap
   - polished side panel list
   - floating CTA report
7. Route `/issues/[id]` sekarang punya wrapper page-level baru:
   - follow card yang lebih jelas
   - timeline activity yang lebih polished
   - aside share/lokasi yang lebih rapi
   - preview lightbox yang konsisten
8. Route `/stats` sekarang card-based penuh:
   - hero statistik
   - filter wilayah polished
   - metric grid
   - status breakdown
   - time stats
   - leaderboard
   - top issue cards

### Admin

9. Route `/admin/login` dipoles ulang dan sekarang punya:
   - `Ingat saya`
   - show/hide password
   - helper keamanan yang menegaskan cookie session tetap server-side
10. Implementasi `Ingat saya` saat ini sengaja aman:
    - hanya menyimpan username admin di localStorage
    - tidak menyimpan password
    - tidak mengubah TTL session backend
11. Shell admin, daftar issue admin, dan detail moderasi admin direfresh agar lebih operasional-friendly dan konsisten dengan design system baru.

### Verification / Docs

12. Verifikasi yang sudah dijalankan:
    - `npm run check`
    - `npm run build`
    - smoke test browser untuk `/`, `/lapor`, `/issues`, `/stats`, `/admin/login`, `/issues/test-issue`
13. Catatan smoke test lokal:
    - error console yang terlihat berasal dari backend lokal/CORS (`localhost:5000`) dan bootstrap device/admin session, bukan dari compile frontend.
14. Diperbarui:
    - `docs/FRONTEND.md`
    - `design-docs/design-system.md`
    - `design-docs/component-spec.md`
    - `design-docs/guide.md`
    - `docs/CHANGELOG_FOR_AGENTS.md`

## 2026-03-20 - Admin Issue Detail Hard-Refresh Loading Fix

- Scope:
  - memperbaiki admin shell yang bisa berhenti di state `Memuat...` saat membuka route detail issue secara initial load/hard refresh.

### Frontend

1. `frontend/src/routes/admin/+layout.svelte` sekarang memakai null-safe access untuk `afterNavigate.from?.url?.pathname`.
2. Pengecekan sesi admin kembali berjalan pada initial navigation ketika objek `from` belum ada, sehingga route `/admin/issues/[id]` tidak macet sebelum `adminMe()` dipanggil.

### Root Cause

3. Implementasi sebelumnya memakai `from?.url.pathname`; pada initial mount, `from` bisa `null`, callback `afterNavigate` melempar runtime error, dan shell admin tertinggal di fallback loading.

### Docs

4. Diperbarui:
   - `docs/FRONTEND.md`
   - `docs/CHANGELOG_FOR_AGENTS.md`

## 2026-03-20 - Deploy Public Smoke Test + Versioned Nginx Template

- Scope:
  - menutup false-green deploy risk tanpa mengganti stack GitHub Actions + VPS + gas build + PM2 + nginx.

### Workflow / Deploy

1. Workflow deploy sekarang membedakan dua tahap verifikasi:
   - runtime lokal di server (`pm2`, port LISTEN, localhost health)
   - smoke test publik via origin/domain/proxy
2. Smoke test publik yang wajib lulus:
   - `${PUBLIC_APP_BASE_URL}/health`
   - `${PUBLIC_API_BASE_URL}/api/v1/health`
   - `${PUBLIC_API_BASE_URL}/api/v1/issues`
   - `${PUBLIC_API_BASE_URL}/api/v1/notifications/stream` dengan expected auth error `401/403`
3. `preflight_frontend` sekarang juga mensyaratkan `PUBLIC_APP_BASE_URL`, bukan hanya `PUBLIC_API_BASE_URL`.
4. Rollback minimum sekarang mengulang verifikasi lokal **dan** smoke test publik; jika public smoke tetap gagal setelah rollback, indikasinya bergeser ke ingress/nginx/domain/runtime config server.

### Runtime Config / Docs

5. Ditambahkan template nginx minimum yang terversioning di repo:
   - `deploy/nginx/jedug.conf.example`
6. `frontend/.env.example` sekarang juga memuat:
   - `PUBLIC_APP_BASE_URL`
7. `docs/DEPLOYMENT.md` diperbarui untuk:
   - flow deploy aktual pasca public smoke
   - source-of-truth template nginx repo
   - checklist manual smoke publik
   - rollback semantics baru
8. `docs/DECISIONS.md` diperbarui untuk mencatat keputusan pemisahan runtime-local vs public-ingress verification.

## 2026-03-20 - Upload Abuse Hardening + Orphan Cleanup

- Scope:
  - menutup blocker abuse pada flow upload/report tanpa mengganti model anonymous device + upload ticket yang sudah live.

### Backend

1. `POST /api/v1/uploads/file/*` sekarang dilindungi rate limit spesifik `10/15m` per-IP.
2. `POST /api/v1/uploads/presign` diperkeras dari sisi service:
   - setiap presign membuat row pending di tabel baru `report_upload_tickets`
   - presign baru ditolak bila device sudah punya terlalu banyak upload pending dalam window pendek (`UPLOAD_PENDING_LIMIT`, default `4`; `UPLOAD_PENDING_WINDOW_SEC`, default `1800`)
3. `UploadService` tidak hanya memverifikasi HMAC `upload_token`, tetapi juga memastikan `object_key` masih punya pending ticket yang dikenal backend dan metadata ticket cocok (`device_id`, `mime_type`, `size_bytes`, expiry).
4. `ReportRepository.SubmitReport` sekarang menghapus row pending upload di transaction yang sama setelah `submission_media` berhasil dipersist, sehingga media yang sudah dipakai report tidak lagi terlihat sebagai orphan.
5. Maintenance runner sekarang juga membersihkan orphan upload:
   - pilih row `report_upload_tickets` yang lebih tua dari `UPLOAD_ORPHAN_RETENTION_SEC` (default `43200`)
   - hapus object storage berdasarkan `upload_mode`
   - baru hapus row ticket dari DB
6. Health snapshot sekarang additive memuat:
   - `upload_orphans_over_retention`
   - `report_upload_tickets_estimate`

### Schema / Config

7. Ditambah migration `backend/migrations/202603200004_create_report_upload_tickets.sql`.
8. Baseline schema dan governance verifier diperbarui agar tabel/index upload ticket menjadi bagian source of truth repo.
9. Env baru:
   - `UPLOAD_PENDING_WINDOW_SEC`
   - `UPLOAD_PENDING_LIMIT`
   - `UPLOAD_ORPHAN_RETENTION_SEC`

### Tests / Docs

10. Ditambah regression test upload service untuk pending upload limit.
11. Diperbarui:
   - `backend/.env.example`
   - `docs/BACKEND.md`
   - `docs/STORAGE_AND_MEDIA.md`
   - `docs/SCHEMA.md`
   - `docs/DEPLOYMENT.md`
   - `docs/DECISIONS.md`
   - `docs/CHANGELOG_FOR_AGENTS.md`

## 2026-03-20 - Report Submit Idempotency Hardening

- Scope:
  - memperbaiki jalur `POST /api/v1/reports` agar retry request yang sama aman dari cooldown tidak semestinya, race retry paralel, dan response replay yang tidak konsisten.
- Akar masalah:
  - service mengecek cooldown sebelum lookup idempotency, sehingga retry request yang sudah sukses tetap bisa berakhir `429`.
  - lookup `client_request_id` masih global, tidak di-scope ke `device_id`.
  - hasil replay tidak konsisten karena backend tidak menyimpan apakah submission awal membuat issue baru atau tidak.
  - retry paralel dengan key yang sama masih mengandalkan race terhadap unique constraint, sehingga salah satu request bisa jatuh ke error internal alih-alih replay hasil lama.
- Perbaikan:
  1. `ReportService` sekarang memvalidasi `client_request_id`, membangun `request_fingerprint`, lalu mengecek replay existing sebelum guard banned/trust/cooldown/upload validation.
  2. `ReportRepository` sekarang mengambil advisory transaction lock ringan berbasis `(device_id, client_request_id)` lalu re-check existing submission di dalam transaction untuk menutup race retry paralel.
  3. `issue_submissions` menyimpan `request_fingerprint` dan `created_issue` agar replay request yang sama mengembalikan `is_new_issue` yang identik dengan submit awal.
  4. Unique constraint idempotency digeser dari global `client_request_id` menjadi composite `(device_id, client_request_id)`.
  5. Reuse `client_request_id` dengan payload fingerprint berbeda sekarang ditolak `409 IDEMPOTENCY_CONFLICT`.
  6. Handler report menambahkan validasi UUID `client_request_id` dan mapping error terstruktur untuk conflict idempotency.
  7. UI `/lapor` menampilkan pesan conflict yang eksplisit bila backend mengembalikan `409`.
- Dampak area:
  - `backend/internal/service/report_service.go`
  - `backend/internal/repository/report_repository.go`
  - `backend/internal/http/handlers/report_handler.go`
  - `backend/internal/service/report_service_test.go`
  - `backend/internal/http/handlers/report_handler_test.go`
  - `backend/internal/service/upload_service_test.go`
  - `backend/migrations/202603200001_harden_report_idempotency.sql`
  - `backend/schema/20260320_000000_baseline.sql`
  - `frontend/src/routes/lapor/+page.svelte`
  - `docs/BACKEND.md`
  - `docs/SCHEMA.md`
  - `docs/FRONTEND.md`
  - `docs/CHANGELOG_FOR_AGENTS.md`
- Action deploy wajib:
  - jalankan migration `backend/migrations/202603200001_harden_report_idempotency.sql` sebelum backend baru menerima traffic penuh.
  - verifikasi unique lama `issue_submissions_client_request_id_key` sudah hilang dan unique baru `(device_id, client_request_id)` aktif.
  - smoke test retry submit yang sama pada koneksi buruk untuk memastikan replay kembali `issue_id/submission_id` yang sama tanpa 429.

## 2026-03-20 - Admin Login Rate Limit + Session Cookie Hardening

- Scope:
  - memperkeras boundary admin auth tanpa mengganti model env credential + in-memory session secara total.

### Backend

1. `POST /api/v1/admin/login` sekarang dilindungi hard rate limit `10/15m` per-IP di router.
2. `AdminService` menambah guard in-memory per fingerprint `ip|username`:
   - window `15 menit`
   - lockout setelah `5` gagal
   - lockout duration `30 menit`
3. Login gagal tetap generic (`username atau password salah`) agar tidak membuka oracle credential.
4. Session admin tetap opaque/random dan in-memory, tetapi login sukses sekarang:
   - merotasi session lama untuk username yang sama
   - mengirim cookie `jedug_admin_session` (`HttpOnly`, `SameSite=Strict`, path `/api/v1/admin`, `Secure` saat production)
5. Ditambah endpoint `POST /api/v1/admin/logout` untuk revoke session server-side + clear cookie.
6. Middleware admin auth sekarang membaca cookie session admin; bearer header lama hanya fallback kompatibilitas.
7. Semua route `/api/v1/admin/*` sekarang diberi header `Cache-Control: no-store`.
8. Config admin diperkeras:
   - `ADMIN_USERNAME` wajib eksplisit
   - `ADMIN_PASSWORD` minimal 12 karakter

### Frontend

9. Helper `src/lib/api/admin.ts` sekarang memakai `credentials: 'include'` untuk seluruh request admin.
10. Admin frontend berhenti menyimpan token auth di `localStorage`.
11. Halaman login dan shell admin sekarang memeriksa sesi lewat `GET /api/v1/admin/me`, dan logout memanggil backend revoke alih-alih sekadar menghapus state client.

### Tests

12. Ditambah regression tests untuk:
   - login cookie set
   - logout revoke + clear cookie
   - middleware auth membaca cookie
   - lockout setelah login gagal berulang
   - rotasi session saat login ulang

### Docs

13. Diperbarui:
   - `backend/.env.example`
   - `docs/BACKEND.md`
   - `docs/FRONTEND.md`
   - `docs/ARCHITECTURE.md`
   - `docs/DEPLOYMENT.md`
   - `docs/MODERATION.md`
   - `docs/DECISIONS.md`
   - `docs/CHANGELOG_FOR_AGENTS.md`

## 2026-03-20 - Notification Auth Transport Hardening

- Scope:
  - memperkecil blast radius token notif tanpa menambah login/account baru dan tanpa membongkar pipeline notif existing.

### Backend

1. `FollowerAuthService` sekarang menerbitkan dua token:
   - `follower_token` untuk endpoint notification/push non-SSE
   - `stream_token` khusus SSE
2. `follower_token` kini device-bound saat verifikasi:
   - token claim membawa `device_token_hash`
   - endpoint non-SSE wajib `X-Follower-Token` + `X-Device-Token`
   - token bocor tanpa device token yang cocok sekarang ditolak
3. `stream_token` diberi purpose `notification_stream` dan TTL pendek agar query string SSE tidak membawa token notif umum.
4. Default TTL notif dipersempit:
   - `FOLLOWER_TOKEN_TTL_SEC` default `43200` (12 jam)
   - `FOLLOWER_STREAM_TOKEN_TTL_SEC` default `600` (10 menit)
5. Handler notification/preferences/push/nearby alerts tetap mempertahankan fallback body/query seperlunya untuk kompatibilitas, tetapi boundary verifikasi final sekarang:
   - non-SSE: token + device token
   - SSE: stream token saja
6. Follow/status response sekarang additive mengembalikan:
   - `follower_stream_token`
   - `follower_stream_token_expires_at`
7. Ditambah regression tests untuk:
   - akses notif dengan device token yang salah
   - token purpose mismatch antara non-SSE vs SSE

### Frontend

8. Helper API notif/push/nearby/preference berhenti mengirim token sensitif di query string untuk jalur non-SSE; transport aktif sekarang memakai header `X-Follower-Token` + `X-Device-Token`.
9. Store auth follower sekarang menyimpan dua cache token:
   - access token non-SSE
   - stream token SSE
10. `notificationsState` sekarang membuka EventSource dengan `stream_token=...`, dan akan refresh stream token secara terpisah saat reconnect.

### Docs

11. Diperbarui:
   - `docs/BACKEND.md`
   - `docs/FRONTEND.md`
   - `docs/DEPLOYMENT.md`
   - `docs/DECISIONS.md`
   - `docs/CHANGELOG_FOR_AGENTS.md`

## 2026-03-20 - Stats Dashboard Scope Correctness + Stable Region Identity

- Scope:
  - menghapus campuran global-vs-scoped pada dashboard `/stats` dan mengunci leaderboard ke identity wilayah yang stabil.

### Backend

1. `GET /api/v1/stats` sekarang memisahkan:
   - `global` sebagai snapshot seluruh issue publik
   - `summary` sebagai totals untuk scope aktif
   - `status` + `time` juga mengikuti scope aktif yang sama
   - `active_scope` sebagai metadata eksplisit (`kind`, `label`, `is_default`)
2. Query summary kini membaca CTE wilayah yang sama dengan leaderboard/top issue, sehingga semua section scoped memakai boundary data yang konsisten.
3. Leaderboard wilayah tidak lagi `GROUP BY name`; grouping sekarang memakai fallback identity stabil:
   - `district_id`
   - lalu `regency_id`
   - lalu `province_id`
4. Row tanpa identity administratif stabil sengaja tidak masuk leaderboard agar ranking tidak misleading.
5. Payload leaderboard kini additive membawa `region_id`, `region_level`, `region_name`, `parent_region_name`, `regency_name`, dan `province_name`.
6. Ditambah regression test ringan untuk helper contract scope stats.

### Frontend

7. `/stats` sekarang merender summary cards dari `summary`, bukan `global`.
8. Copy halaman menegaskan bahwa ringkasan, status, time stats, leaderboard, dan top issue memakai scope aktif yang sama.
9. Leaderboard keyed dengan `region_id`, bukan `district_name`, dan menampilkan konteks parent administratif agar nama wilayah yang sama tidak terlihat ambigu.

### Docs

10. Diperbarui:
   - `docs/BACKEND.md`
   - `docs/FRONTEND.md`
   - `design-docs/guide.md`
   - `docs/DECISIONS.md`
   - `docs/CHANGELOG_FOR_AGENTS.md`

## 2026-03-20 - Moderation Correctness Boundary Hardening

- Scope:
  - menutup false success di moderation admin dan memperjelas boundary antara perubahan domain inti vs audit/event.

### Backend

1. `hide/unhide issue` dan `ban device` sekarang mengecek `RowsAffected`; target yang tidak ada dikembalikan sebagai `404` lewat sentinel error yang konsisten.
2. `fix/reject issue` sekarang memakai satu transaksi domain untuk:
   - lock issue target
   - fail eksplisit jika issue tidak ada
   - update `issues.status`
   - adjust trust score submitter hanya bila status benar-benar berubah
3. `status_updated` event tidak lagi diinsert di dalam transaksi utama moderation; event dipublish sesudah commit sebagai best-effort.
4. `moderation_actions` juga dipindahkan ke post-commit best-effort untuk menghindari `500` palsu setelah action utama sebenarnya sudah committed.
5. Repeated `fix/reject` pada status yang sama sekarang tidak lagi menggandakan trust adjustment atau `status_updated` event.
6. Ditambah regression tests untuk:
   - mapping `404` pada admin handler
   - audit/event failure yang tidak lagi memblokir success path action utama

### Docs

7. Memperbarui `docs/BACKEND.md` untuk menjelaskan semantics moderation baru: not-found handling, transaksi domain inti, dan audit post-commit.

## 2026-03-20 - Browser Push Delivery Durability via DB Outbox

- Scope:
  - mengurangi loss pada browser push async tanpa menambah broker eksternal atau sistem streaming besar.

### Backend

1. Browser push tidak lagi hanya mengandalkan queue in-memory; `DeliverBatch(...)` sekarang menulis batch ke tabel outbox `push_delivery_jobs`.
2. Ditambahkan repository baru `pushDeliveryJobRepository` untuk:
   - enqueue idempotent per `(event_id, follower_id)`
   - claim batch siap kirim via `FOR UPDATE SKIP LOCKED`
   - mark `delivered`, `retry`, atau `failed`
3. Worker push sekarang memproses outbox DB dengan poll + wake ringan, sehingga restart/crash tidak menghilangkan job yang belum selesai.
4. Retry behavior dibuat eksplisit:
   - maksimum `5` attempt
   - backoff `30s -> 2m -> 5m -> 15m`
   - `429` dan `5xx` dianggap retryable
   - `404/410` men-disable subscription dan tidak di-retry
5. Job failure sekarang tercatat di DB lewat `attempt_count`, `next_attempt_at`, `delivered_at`, `failed_at`, dan `last_error`, sehingga loss tidak lagi hanya terlihat di log.
6. Notification dispatch biasa dan nearby alert sekarang selalu membawa `event_id` ke jalur push agar dedupe outbox stabil.
7. Ditambahkan regression tests untuk enqueue outbox, retry/fail boundary, dan handling job tanpa subscription aktif.

### Schema

8. Ditambahkan migration `backend/migrations/202603200002_create_push_delivery_jobs.sql`.
9. Baseline `backend/schema/20260320_000000_baseline.sql` dan governance script schema diperbarui agar fresh bootstrap dan verify repo-aware tetap lulus.

### Docs

10. Diperbarui:
   - `docs/BACKEND.md`
   - `docs/SCHEMA.md`
   - `docs/ARCHITECTURE.md`
   - `docs/DECISIONS.md`
   - `docs/CHANGELOG_FOR_AGENTS.md`

## 2026-03-20 - Observability + Retention Minimum untuk Runtime Data

- Scope:
  - menutup gap observability dasar dan retention minimum tanpa menambah monitoring stack besar atau broker baru.

### Backend

1. `GET /api/v1/health` sekarang mengembalikan snapshot ringan berisi:
   - uptime
   - DB check
   - SSE follower/connection count + cumulative dropped frames
   - push queue ready + failed 24 jam terakhir
   - retention debt (`notifications`, stale/disabled `push_subscriptions`)
   - estimasi growth `issue_events`, `notifications`, `push_subscriptions`, `push_delivery_jobs`
2. Ditambahkan package `internal/ops` untuk:
   - health snapshot query
   - retention cleanup query
   - runner periodik in-process dengan advisory lock DB
3. Runner retention default:
   - interval `6 jam`
   - delete `notifications` > `90 hari`
   - soft-disable active `push_subscriptions` yang stale > `180 hari`
   - delete disabled `push_subscriptions` > `30 hari`
   - delete `push_delivery_jobs.delivered` > `14 hari`
   - delete `push_delivery_jobs.failed` > `30 hari`
4. Ditambahkan command manual `go run ./cmd/maintenance` dan target `make ops-retention`.
5. `issue_events` belum di-prune otomatis; policy saat ini adalah keep + observe karena tabel ini masih dipakai sebagai timeline publik/audit.
6. Logging penting diperjelas:
   - report failure sekarang menyertakan request id
   - admin login/moderation success/failure lebih eksplisit
   - SSE stream open/close sekarang di-log dengan replay count dan close reason

### Schema / Ops

7. Ditambahkan migration `backend/migrations/202603200003_add_retention_indexes.sql` untuk query retention/health yang lebih murah:
   - `notifications.created_at`
   - `push_subscriptions.disabled_at`
   - `push_subscriptions.updated_at` (active partial)
   - `push_delivery_jobs.delivered_at`
   - `push_delivery_jobs.failed_at`
8. Baseline schema dan governance script diperbarui agar fresh bootstrap dan verify tetap sinkron.
9. Saat patch ini dibuat, deploy workflow sempat menegakkan `pm2-logrotate`; per 2026-03-20 langkah itu sudah dikeluarkan dari critical deploy path dan dipindah menjadi setup manual satu kali di VPS.

### Docs

10. Diperbarui:
   - `docs/BACKEND.md`
   - `docs/SCHEMA.md`
   - `docs/ARCHITECTURE.md`
   - `docs/DEPLOYMENT.md`
   - `docs/DECISIONS.md`
   - `docs/CHANGELOG_FOR_AGENTS.md`

## 2026-03-20 - Schema Governance Baseline + Migration Repo

- Scope:
  - menjadikan schema database reproducible langsung dari repo tanpa bergantung pada file SQL eksternal.

### Backend / DB

1. Menambah baseline penuh `backend/schema/20260320_000000_baseline.sql` yang merepresentasikan schema JEDUG saat ini, termasuk extension `postgis` dan `pgcrypto`.
2. Membuat folder migration nyata `backend/migrations/` beserta file additive/idempotent untuk:
   - `issue_events`
   - `submission_media`
   - `issue_followers`
   - `notifications`
   - `push_subscriptions`
   - `follower_auth_bindings`
   - `notification_preferences`
   - `nearby_alerts`
3. Menambah helper operasional:
   - `backend/scripts/bootstrap_db.sh`
   - `backend/scripts/verify_schema_governance.sh`
4. Menambah target Makefile:
   - `make db-bootstrap`
   - `make db-upgrade`
   - `make db-verify-schema`

### Docs

5. Menyinkronkan source of truth schema di:
   - `docs/SCHEMA.md`
   - `docs/BACKEND.md`
   - `docs/ARCHITECTURE.md`
   - `docs/DEPLOYMENT.md`
   - `docs/DECISIONS.md`
   - `backend/README.md`
   - `AGENTS.md`
6. Mismatch historis yang sengaja dicatat, bukan diubah diam-diam:
   - file SQL eksternal lama memiliki typo `submission_media.widthINT/heightINT`
   - sebagian query backend masih defensif terhadap status issue `verified` / `in_progress`, tetapi baseline schema tetap memakai enum issue kanonik yang sudah ada
   - `notifications.event_id` tetap logical reference ke `issue_events.id` tanpa FK untuk rollout-safe compatibility
7. Script DB helper sekarang juga auto-load `backend/.env` bila `DATABASE_URL` belum diexport di shell, agar `make db-bootstrap` / `make db-upgrade` / `make db-verify-schema` langsung usable di workflow dev biasa.

## 2026-03-20 - Hardening Upload Publik dengan Device-Bound Upload Ticket

- Scope:
  - menutup jalur upload publik yang sebelumnya cukup mengandalkan `object_key`/presign tanpa binding kuat ke device dan report flow.

### Backend

1. `POST /api/v1/uploads/presign` sekarang wajib `anon_token` yang valid.
2. Backend menerbitkan `upload_token` HMAC berttl pendek (`UPLOAD_TICKET_TTL_SEC`, default 10 menit) untuk kombinasi:
   - `device_id`
   - `object_key`
   - `mime_type`
   - `size_bytes`
   - purpose `report_media`
3. `POST /api/v1/uploads/file/{object_key}` sekarang wajib header `X-Upload-Token`; upload local tanpa proof ini ditolak.
4. `POST /api/v1/reports` sekarang menuntut `media[].upload_token` dan memverifikasi:
   - ticket valid / belum expired
   - `device_id` submit cocok dengan owner ticket
   - `object_key`, `mime_type`, `size_bytes` cocok dengan ticket
   - object benar-benar ada di storage aktif
   - `object_key` belum pernah dipakai di `submission_media`
5. Storage abstraction sekarang punya `Stat(...)` agar submit report bisa mengecek keberadaan object di local maupun R2 sebelum report diterima.
6. Env backend baru:
   - `UPLOAD_TOKEN_SECRET` (opsional; default fallback ke `FOLLOWER_TOKEN_SECRET`)
   - `UPLOAD_TICKET_TTL_SEC`

### Frontend

7. Flow `/lapor` kini:
   - kirim `anon_token` saat presign
   - kirim `X-Upload-Token` saat upload local/fallback local
   - kirim `media[].upload_token` saat submit report

### Doc Updates

- `docs/BACKEND.md`
- `docs/DEPLOYMENT.md`
- `docs/STORAGE_AND_MEDIA.md`
- `docs/FRONTEND.md`
- `docs/DECISIONS.md`
- `docs/CHANGELOG_FOR_AGENTS.md`

## 2026-03-16 - Fix Filter Statistik Wilayah Administratif

- Scope:
  - memperbaiki dropdown provinsi/kabupaten-kota di `/stats` agar tidak lagi bergantung penuh pada payload `/stats` yang sama.
  - memperkuat penurunan wilayah administratif backend untuk issue lama/new reports supaya leaderboard dan top issue benar-benar berbasis wilayah administratif.
- Dampak area:
  - `backend/internal/domain/stats.go`
  - `backend/internal/repository/stats_repository.go`
  - `backend/internal/repository/report_repository.go`
  - `backend/internal/repository/location_repository.go`
  - `backend/internal/service/stats_service.go`
  - `backend/internal/service/stats_service_test.go`
  - `backend/internal/service/location_service.go`
  - `backend/internal/http/handlers/stats.go`
  - `backend/internal/http/router.go`
  - `frontend/src/lib/api/stats.ts`
  - `frontend/src/lib/api/types.ts`
  - `frontend/src/routes/stats/+page.svelte`
- File docs yang diupdate:
  - `docs/BACKEND.md`
  - `docs/FRONTEND.md`
  - `design-docs/guide.md`
  - `docs/CHANGELOG_FOR_AGENTS.md`
- Catatan penting untuk agent berikutnya:
  - endpoint baru `GET /api/v1/stats/regions/options` sekarang menjadi sumber utama dropdown wilayah di frontend.
  - query stats tidak lagi hanya bergantung pada `issues.region_id`; ia fallback ke `latest issue_submissions.region_id` lalu spatial lookup dari `issues.public_location`.
  - normalisasi level wilayah kini mengenali alias English + Indonesia (`provinsi`, `kabupaten`, `kota`, `kecamatan`) agar data `regions` dengan label berbeda tetap cocok untuk filter stats dan default geolocation.
  - `GET /api/v1/location/label` kini additive membawa `district_name`, `regency_name`, dan `province_name` untuk membantu matching default scope `/stats`.

## 2026-03-16 - iPhone Push UX + Stats Wilayah Administratif

- Scope:
  - memperbaiki UX browser push di iPhone/iOS agar membedakan Safari tab biasa vs Home Screen app.
  - merombak `/stats` supaya leaderboard + top issue berbasis wilayah administratif (`provinsi` / `kabupaten-kota` / `kecamatan`), bukan fallback `Sekitar Jalan ...`.
- Dampak area:
  - `frontend/src/lib/stores/browser-push.ts`
  - `frontend/src/lib/components/BrowserPushCard.svelte`
  - `frontend/src/lib/components/NotificationPreferencesPanel.svelte`
  - `frontend/src/lib/utils/follower-auth.ts`
  - `frontend/src/lib/utils/storage.ts`
  - `frontend/src/lib/stores/notification-preferences.ts`
  - `backend/internal/domain/stats.go`
  - `backend/internal/repository/stats_repository.go`
  - `backend/internal/service/stats_service.go`
  - `backend/internal/service/stats_service_test.go`
  - `backend/internal/http/handlers/stats.go`
  - `frontend/src/lib/api/stats.ts`
  - `frontend/src/lib/api/types.ts`
  - `frontend/src/routes/stats/+page.svelte`
- File docs yang diupdate:
  - `docs/BACKEND.md`
  - `docs/FRONTEND.md`
  - `design-docs/component-spec.md`
  - `design-docs/guide.md`
  - `docs/CHANGELOG_FOR_AGENTS.md`
- Mismatch baru (jika ada):
  - geolocation default di `/stats` masih mengandalkan `GET /api/v1/location/label` + pencocokan nama region, jadi akurasi fallback terbaik terjadi saat data `regions` internal tersedia dan konsisten.
  - recovery `follower_binding_not_found` saat ini di-handle di frontend dengan reset identitas anonim lokal + reload; belum ada endpoint backend khusus untuk repair tanpa reset storage.

## 2026-03-16 - Nearby Alerts untuk Area Pantauan Lokal

- Scope: menambah watched locations anonim agar follower bisa menerima notifikasi issue baru di sekitar area pilihan tanpa follow issue satu per satu.

### Backend

1. Menambah migration `backend/migrations/202603160004_create_nearby_alerts.sql` untuk:

- tabel `nearby_alert_subscriptions`
- tabel `nearby_alert_deliveries`
- kolom preference baru `notification_preferences.notify_on_nearby_issue_created`

2. Menambah endpoint:

- `GET /api/v1/nearby-alerts?follower_token=...`
- `POST /api/v1/nearby-alerts`
- `PATCH /api/v1/nearby-alerts/:id`
- `DELETE /api/v1/nearby-alerts/:id`

3. Nearby Alerts memakai auth yang sama dengan notif/push existing:

- `follower_id` tetap browser-scoped
- ownership diverifikasi via `follower_token`
- tidak ada login/account baru

4. Guard service minimum:

- maksimum `10` lokasi pantauan per follower/browser
- radius valid `100..5000m`
- patch koordinat harus mengirim latitude/longitude berpasangan

5. Dispatch nearby alert sekarang di-hook dari `report_repository.insertTimelineEvents()` hanya saat event `issue_created`.
6. `DispatchNearbyAlertsForIssueCreated(...)` melakukan:

- lookup subscription `enabled` via `ST_DWithin`
- insert dedupe row ke `nearby_alert_deliveries` (`UNIQUE(subscription_id, issue_id)`)
- group hasil per follower agar overlap beberapa lokasi pantauan tidak menghasilkan notif ganda
- create in-app notification type `nearby_issue_created`
- enqueue browser push bila channel push aktif

7. Self-notify prevention ikut berlaku: reporter yang sama (`actor_follower_id`) tidak menerima nearby alert untuk issue baru yang ia buat sendiri.

### Frontend

8. Menambah helper API `frontend/src/lib/api/nearby-alerts.ts`.
9. Menambah store lazy `frontend/src/lib/stores/nearby-alerts.ts`.
10. Menambah `NearbyAlertsPanel.svelte` di notification center header, dengan UX minimum:

- tambah lokasi pantauan
- autofill lokasi browser opsional
- input manual latitude/longitude
- edit label + radius
- aktif/nonaktifkan
- hapus lokasi pantauan

11. Panel preferensi notifikasi kini juga menambah toggle event `notify_on_nearby_issue_created`.

### Doc Updates

- `docs/BACKEND.md`
- `docs/FRONTEND.md`
- `docs/SCHEMA.md`
- `docs/DECISIONS.md`
- `design-docs/component-spec.md`
- `design-docs/guide.md`

## 2026-03-16 - Notification Preferences for Anonymous Followers

- Scope: menambah kontrol minimum agar follower anonim bisa mengatur channel dan jenis update notifikasi tanpa merombak arsitektur notif yang sudah ada.

### Backend

1. Menambah tabel `notification_preferences` via `backend/migrations/202603160003_create_notification_preferences.sql`.
2. Menambah endpoint:
   - `GET /api/v1/notification-preferences?follower_token=...`
   - `PATCH /api/v1/notification-preferences`
3. Preference tetap memakai auth yang sama dengan notif/push sekarang:
   - `follower_token`
   - `follower_id` tidak diterima sebagai bearer secret untuk update settings
4. `DispatchNotificationsForEvent` sekarang menjadi titik integrasi preference:
   - filter global `notifications_enabled`
   - filter per-event (`photo/status/severity/casualty`)
   - filter channel `in_app_enabled`
   - filter channel `push_enabled`
5. In-app notification dan browser push tidak lagi dikunci ke keputusan channel yang sama:
   - in-app bisa off, push tetap on
   - push bisa off, in-app tetap on
6. `GET /notification-preferences` mengembalikan default sintetis; row DB baru dibuat saat user pertama kali menyimpan preference. `push_enabled` default `true` hanya bila follower memang sudah punya subscription push aktif saat itu.

### Frontend

7. Menambah helper API `frontend/src/lib/api/notification-preferences.ts`.
8. Menambah store `frontend/src/lib/stores/notification-preferences.ts`.
9. Notification center header sekarang punya panel ringan `Preferensi Notifikasi`:
   - master switch
   - toggle in-app
   - toggle push
   - toggle per event type
10. Toggle push dibuat state-aware:

- jika browser push belum aktif, user diarahkan ke `BrowserPushCard`
- jika push sudah aktif, toggle preference bisa diubah tanpa halaman settings baru

11. Update preference disiarkan antar tab via `BroadcastChannel` + fallback `storage` event.
12. Jika in-app dimatikan, `notificationsState` memutus koneksi SSE realtime agar browser tidak mempertahankan stream yang tidak dipakai.

### Doc Updates

- `docs/BACKEND.md`
- `docs/FRONTEND.md`
- `docs/SCHEMA.md`
- `design-docs/component-spec.md`
- `design-docs/guide.md`

## 2026-03-16 - Browser Web Push Notifications for Followed Issues

- Scope: menambah channel browser push di atas notifikasi in-app yang sudah ada, lengkap dengan service worker, subscription storage, dan delivery backend.

### Backend

1. Menambah tabel `push_subscriptions` via `backend/migrations/202603160001_create_push_subscriptions.sql`.
2. Menambah endpoint:
   - `GET /api/v1/push/status?follower_id=...`
   - `POST /api/v1/push/subscribe`
   - `POST /api/v1/push/unsubscribe`
3. Menambah `internal/push/notifier.go` berbasis `github.com/SherClockHolmes/webpush-go` dengan payload minimal `title/body/issue_id/url/type`.
4. `DispatchNotificationsForEvent` sekarang juga memanggil push notifier batch setelah row `notifications` berhasil dibuat; SSE tetap dipertahankan.
5. Endpoint push yang `404/410` dari provider Web Push otomatis menandai row subscription sebagai `disabled_at`.
6. Config backend baru:
   - `WEB_PUSH_VAPID_PUBLIC_KEY`
   - `WEB_PUSH_VAPID_PRIVATE_KEY`
   - `WEB_PUSH_SUBSCRIBER`
   - `WEB_PUSH_SITE_URL`
   - `WEB_PUSH_TTL_SEC`
7. Startup backend fail-fast jika env Web Push hanya terisi sebagian.

### Frontend

8. Menambah store `frontend/src/lib/stores/browser-push.ts` untuk state:
   - `unsupported`
   - `default`
   - `granted`
   - `denied`
   - `subscribed`
9. Menambah helper API `frontend/src/lib/api/push.ts`.
10. Menambah service worker `frontend/static/sw.js` dan icon `frontend/static/push-icon.svg`.
11. `routes/+layout.svelte` sekarang menginisialisasi browser push setelah bootstrap device + notification store.
12. `AppHeader.svelte` menampilkan CTA ringan browser push di panel notifikasi.
13. `routes/issues/[id]/+page.svelte` menampilkan CTA browser push setelah browser anonim mengikuti issue.
14. Service worker behavior:
    - tab visible -> kirim `postMessage` untuk refresh issue lokal, tanpa OS notification ganda
    - tab tidak aktif -> tampilkan browser notification
    - klik notification -> fokus/navigate ke issue terkait

### Doc Updates

- `docs/BACKEND.md`
- `docs/FRONTEND.md`
- `docs/SCHEMA.md`
- `docs/DEPLOYMENT.md`
- `docs/DECISIONS.md`
- `design-docs/component-spec.md`
- `design-docs/guide.md`

## 2026-03-16 - Notification Hardening: Follower Token Auth, Push SSRF Guard, Async Delivery

- Scope: menutup temuan audit P1/P2 utama pada notifikasi tanpa menambah login penuh.

### Backend

1. Menambah tabel `follower_auth_bindings` via `backend/migrations/202603160002_create_follower_auth_bindings.sql`.
2. Menambah endpoint `POST /api/v1/followers/auth` untuk refresh `follower_token` memakai `X-Device-Token`.
3. Follow/status kini mengikat `follower_id` ke browser anonim yang sama dan mengembalikan `follower_token` + `follower_token_expires_at`.
4. Endpoint notification dan push kini memakai `follower_token`, bukan `follower_id` mentah:
   - `GET /api/v1/notifications?follower_token=...`
   - `PATCH /api/v1/notifications/:id/read?follower_token=...`
   - `DELETE /api/v1/notifications/:id?follower_token=...`
   - `GET /api/v1/notifications/stream?follower_token=...`
   - `GET /api/v1/push/status?follower_token=...`
   - `POST /api/v1/push/subscribe`
   - `POST /api/v1/push/unsubscribe`
5. `PushService` sekarang memvalidasi endpoint Web Push via allowlist host/path + wajib HTTPS + reject credential/IP host + validasi panjang key `p256dh/auth` untuk menutup SSRF endpoint arbitrary.
6. `internal/push/notifier.go` sekarang memakai queue in-process + worker goroutine agar submit report / moderasi tidak menunggu delivery push selesai.

### Frontend

7. Menambah helper `frontend/src/lib/utils/follower-auth.ts` dan API `frontend/src/lib/api/follower-auth.ts`.
8. Notification store sekarang:
   - refresh `follower_token` bila perlu
   - memperbaiki reconnect SSE (tidak hard-stop, timer dibersihkan sebelum reconnect)
   - melakukan fallback refresh ringan saat gagal berulang
   - sinkron read/delete antar tab via `BroadcastChannel` + `storage` event
9. Browser push store sekarang memakai `follower_token` untuk status/subscribe/unsubscribe dan menolak enable bila browser belum punya binding follower yang sah.

### Env / Docs

10. Env backend baru:
    - `FOLLOWER_TOKEN_SECRET`
    - `FOLLOWER_TOKEN_TTL_SEC`
11. Update docs:
    - `docs/BACKEND.md`
    - `docs/FRONTEND.md`
    - `docs/SCHEMA.md`
    - `docs/DEPLOYMENT.md`
    - `docs/DECISIONS.md`

## 2026-03-15 - Notification Delete UX + Smart Same-Issue Refresh

- Scope: menambah delete notification end-to-end dan membuat klik notification pada issue aktif melakukan refresh lokal, bukan navigate sia-sia.

### Backend

1. Menambah endpoint `DELETE /api/v1/notifications/:id?follower_id=...`.
2. Authorization delete mengikuti pasangan `notification_id + follower_id`; row milik follower lain tidak terhapus.
3. Delete dibuat aman/idempotent dengan payload `deleted: true|false`.
4. `PATCH /api/v1/notifications/:id/read` kini mengembalikan `data.read_at` dari DB agar frontend bisa sinkron tanpa timestamp lokal palsu.

### Frontend

5. `stores/notifications.ts` menambah state `deletingIDs` + method `delete(notificationID)` untuk removal lokal ringan tanpa full reload.
6. `AppHeader.svelte` menambah action hapus per item dan logic route-aware:
   - issue berbeda -> `goto('/issues/{id}')`
   - issue sama -> dispatch refresh event lokal
7. Menambah helper `frontend/src/lib/utils/issue-detail-refresh.ts` untuk event ringan refresh detail issue.
8. `routes/issues/[id]/+page.svelte` kini:
   - listen refresh event dari notification click
   - re-fetch detail issue, timeline, dan follow state saat issue sama
   - menampilkan micro-feedback `Laporan diperbarui`
   - sinkron lagi ke `data` route saat user pindah ke issue lain di route yang sama

### Doc Updates

- `docs/BACKEND.md`
- `docs/FRONTEND.md`
- `design-docs/component-spec.md`
- `design-docs/guide.md`

## 2026-03-15 - Realtime In-App Notifications via SSE

- Scope: SSE backend hub + endpoint + frontend EventSource client dengan exponential-backoff reconnect.

## 2026-03-15 - Notification Business Rules + Stats Fallback Fixes

- Scope: mencegah self-notify, memastikan read/unread persisten benar, memperbaiki fallback wilayah stats, dan memperjelas copy notifikasi.

### Backend

1. **Self-notify prevention**

- `report` submit kini menerima `actor_follower_id` (opsional UUID).
- `DispatchNotificationsForEvent` menambah parameter `excludeFollowerID *uuid.UUID`.
- Query insert notifikasi men-skip follower actor: `AND ($6::uuid IS NULL OR follower_id <> $6::uuid)`.

2. **Contextual notification copy**

- Title/message builder di `notification_repository.go` kini menyertakan label lokasi issue.
- Prioritas label: `road_name` → region issue → region submission terbaru → `Issue #<short-id>`.

3. **Read/unread correctness**

- `MarkAsRead` repository kini mengembalikan `(updated bool, err error)` memakai `RowsAffected()`.
- `PATCH /notifications/:id/read` mengembalikan 404 jika row not found atau follower tidak cocok (tidak lagi sukses palsu).

4. **Stats fallback lebih manusiawi**

- Query leaderboard dan top issue kini fallback region dari `latest issue_submissions.region_id` jika `issues.region_id` kosong.
- Fallback final jadi `Sekitar {road_name}` lalu `Area Lainnya` (mengurangi label `Wilayah Tidak Diketahui`).

### Frontend

5. `lapor/+page.svelte` mengirim `actor_follower_id` saat submit report agar backend bisa skip self-notify.
6. `stores/notifications.ts` menambah error handling mark-read: jika gagal sinkron, tampilkan error ringan dan refresh list dari server agar unread badge tetap konsisten dengan DB.
7. `lib/api/types.ts` menambah field `actor_follower_id` di `ReportInput`.

### Doc Updates

- `docs/BACKEND.md`
- `docs/FRONTEND.md`
- `docs/SCHEMA.md`

---

### Backend

1. **`internal/sse/hub.go`** (baru): `Hub` struct dengan `Subscribe(followerID) (ch, done)` dan `Push(followerID, msg)`. `sse.Default` adalah singleton global. Thread-safe (RWMutex). Buffer per-connection 16 slot, non-blocking drop.
2. **`notification_repository.go`**: `DispatchNotificationsForEvent` diubah dari `db.Exec` → `db.Query` dengan `RETURNING id, issue_id, follower_id, type, title, message, created_at`. Setiap row yang baru di-insert dipush ke `sse.Default.Push(followerID, sseMsg)` dengan format `event: notification\ndata: {...}\n\n`.
3. **`notification_handler.go`**: Tambah handler `Stream` untuk SSE (`GET /api/v1/notifications/stream`). Menggunakan `fasthttp.StreamWriter` + `bufio.Writer`. Heartbeat ping setiap 30 detik. Disconnect detection via Flush error. Header `X-Accel-Buffering: no` untuk nginx.
4. **`router.go`**: Route `api.Get("/notifications/stream", notifHandler.Stream)` ditambahkan sebelum route `:id/read`.

### Frontend

5. **`stores/notifications.ts`**: Tambah SSE client module-level (`_connectSSE`, `_disconnectSSE`). SSE dibuka setelah `refresh()` berhasil fetch initial state. Event `notification` di-prepend ke items (deduplication by id). Reconnect dengan exponential backoff (1s → 30s cap, max 10 attempt). EventSource tidak akan dibuka di SSR (`typeof EventSource === 'undefined'` guard). `notificationsState.disconnect` tersedia untuk cleanup eksplisit jika diperlukan.

### Nginx

6. Blok `location /api/v1/notifications/stream` dengan `proxy_buffering off`, `proxy_read_timeout 3600s`, `proxy_http_version 1.1` wajib ditambahkan di nginx server. Lihat `docs/DEPLOYMENT.md` untuk config lengkap.

### CI/CD

7. Tidak ada perubahan workflow deploy yang diperlukan. Nginx config diapply manual sekali di VPS.

### Doc Updates

- `docs/BACKEND.md`: endpoint SSE + hub description
- `docs/DEPLOYMENT.md`: blok nginx SSE config + guidance CI/CD

---

## 2026-03-15 - In-App Notification UI + Mark-as-Read Endpoint

- Scope: menyelesaikan alur notifikasi in-app end-to-end di frontend, plus endpoint mark-as-read minimal di backend.
- Kondisi awal:
  - pipeline backend `issue_events -> notifications` sudah berjalan,
  - tetapi frontend belum pernah memanggil `/api/v1/notifications`, sehingga badge/panel notifikasi tidak muncul.
- Perbaikan backend:
  1. `NotificationRepository` menambah `MarkAsRead(notificationID, followerID)`.
  2. `NotificationService` menambah method `MarkAsRead`.
  3. `NotificationHandler` menambah handler `PATCH /api/v1/notifications/:id/read?follower_id=...`.
  4. Router menambah route patch tersebut.
- Perbaikan frontend:
  1. API client menambah helper `apiPatch`.
  2. API notifications menambah `markNotificationRead(notificationID, followerID)`.
  3. Menambah store terpusat `frontend/src/lib/stores/notifications.ts` untuk fetch/list/mark-read + unread count.
  4. `routes/+layout.svelte` memanggil `notificationsState.init()` saat app load (setelah bootstrap device).
  5. `AppHeader.svelte` menambah UI lonceng, unread badge, dan dropdown panel notifikasi.
  6. Klik item notifikasi: mark as read lalu redirect ke `/issues/{issue_id}`.
- Verifikasi lokal:
  - `go test ./...` ✅
  - `npm run check` ✅ (0 errors, 0 warnings)

## 2026-03-15 - Notification Bug Fix: int64 vs UUID mismatch on issue_events.id

- Scope: fix scan error `unable to scan type int64 into UUID` yang menyebabkan event insert gagal dan notifikasi tidak terkirim.
- Akar masalah:
  - `issue_events.id` di production DB adalah `BIGSERIAL` (int64).
  - Kode yang dibuat di task sebelumnya mengasumsikan kolom ini UUID → `RETURNING id` di-scan ke `uuid.UUID` → scan error → event dianggap gagal → notifikasi tidak dibuat.
- Perbaikan (minimal, follow actual schema):
  1. `domain/notification.go`: `EventID uuid.UUID` → `EventID int64`.
  2. `repository/notification_repository.go`: signature `DispatchNotificationsForEvent(..., eventID uuid.UUID, ...)` → `eventID int64`.
  3. `repository/report_repository.go`: `var eventID uuid.UUID` → `var eventID int64`; format log `%s` → `%d`.
  4. `repository/admin_repository.go`: `eventID uuid.UUID` → `eventID int64`; format log `%s` → `%d`.
  5. `migrations/202603150001_create_notifications.sql`: `event_id UUID` → `event_id BIGINT`.
  6. `frontend/src/lib/api/notifications.ts`: `event_id: string` → `event_id: number`.
- Tidak ada perubahan ke `issue_events` schema — sengaja mengikuti schema yang sudah ada.
- Action wajib di prod: jalankan migration `202603150001_create_notifications.sql` jika belum.

## 2026-03-15 - Follow Notification Pipeline: issue_events → notifications

- Scope: backend notification pipeline + GET `/api/v1/notifications` endpoint + frontend API helper.
- Akar masalah yang ditemukan:
  - `issue_events` sudah menyimpan event (issue_created, photo_added, severity_changed, casualty_reported, status_updated) — benar.
  - `issue_followers` sudah menyimpan data follower — benar.
  - Tetapi tidak ada kode yang menghubungkan keduanya; tidak ada tabel `notifications`, tidak ada dispatch logic.
  - `issue_timeline` tidak pernah ada di codebase (tidak ada referensi, tidak ada tabel).
- Perbaikan:
  1. Membuat tabel `notifications` via `backend/migrations/202603150001_create_notifications.sql` — **WAJIB DIJALANKAN DI PROD**.
  2. Membuat `backend/internal/domain/notification.go` — struct `Notification`.
  3. Membuat `backend/internal/repository/notification_repository.go` — interface `NotificationRepository` + free-function `DispatchNotificationsForEvent(ctx, db, issueID, eventID, eventType)` yang dipakai bersama oleh dua repo.
  4. Membuat `backend/internal/service/notification_service.go` — `NotificationService.GetByFollowerID`.
  5. Membuat `backend/internal/http/handlers/notification_handler.go` — `GET /api/v1/notifications?follower_id=<uuid>&limit=<n>`.
  6. Mengubah `report_repository.insertTimelineEvents`: INSERT `issue_events` kini memakai `RETURNING id`; setelah berhasil, memanggil `DispatchNotificationsForEvent` (non-fatal).
  7. Mengubah `admin_repository.UpdateIssueStatus`: INSERT `issue_events` kini memakai `RETURNING id`; dispatch notifikasi dipanggil setelah `tx.Commit()` berhasil (non-fatal).
  8. Mengubah `router.go`: wire `notifRepo`, `notifSvc`, `notifHandler`; menambah route `GET /api/v1/notifications`.
  9. Membuat `frontend/src/lib/api/notifications.ts` — helper `getNotifications(followerID, limit)`.
- Event yang memicu notifikasi:
  - `issue_created`, `photo_added`, `severity_changed`, `casualty_reported` → dari submit report
  - `status_updated` → dari admin moderation action
- Endpoint final:
  - `GET /api/v1/notifications?follower_id=<uuid>[&limit=50]`
  - Response: `{success: true, data: {items: [{id, issue_id, event_id, type, title, message, created_at},...]}}`
- Deduplication: `UNIQUE(event_id, follower_id)` + `ON CONFLICT DO NOTHING` — dispatch idempotent.
- Dampak area:
  - `backend/migrations/202603150001_create_notifications.sql` (baru — **WAJIB RUN DI PROD**)
  - `backend/internal/domain/notification.go` (baru)
  - `backend/internal/repository/notification_repository.go` (baru)
  - `backend/internal/service/notification_service.go` (baru)
  - `backend/internal/http/handlers/notification_handler.go` (baru)
  - `backend/internal/repository/report_repository.go` (diubah — INSERT RETURNING + dispatch)
  - `backend/internal/repository/admin_repository.go` (diubah — INSERT RETURNING + dispatch + `log` import)
  - `backend/internal/http/router.go` (diubah — wire notif + route)
  - `frontend/src/lib/api/notifications.ts` (baru)
- Action wajib setelah deploy:
  - Jalankan: `psql "$DATABASE_URL" -f migrations/202603150001_create_notifications.sql`
  - Verifikasi: submit report baru → cek tabel `notifications` untuk follower issue tersebut.
  - Verifikasi: admin ubah status issue → cek tabel `notifications` untuk follower issue tersebut.

## 2026-03-15 - CI/CD Hotfix: PM2 bootstrap for non-interactive SSH

- Scope:
  - memperbaiki kegagalan deploy CI saat `gas build` berjalan di sesi SSH non-interactive dan PM2 tidak ditemukan di PATH.
  - menambah langkah preflight untuk memastikan PM2 siap sebelum deploy backend/frontend.
- Perbaikan:
  - workflow kini mencoba menambahkan PATH Node dari lokasi NVM umum (`~/.nvm/versions/node/*/bin`).
  - bila `pm2` belum tersedia, workflow install PM2 user-local via `npm install -g pm2 --prefix ~/.local` lalu lanjut deploy.
  - jika npm juga tidak tersedia, workflow gagal cepat dengan pesan error eksplisit.
- Dampak area:
  - `.github/workflows/deploy.yml`
  - `docs/DEPLOYMENT.md`
  - `docs/CHANGELOG_FOR_AGENTS.md`
- Mismatch baru (jika ada):
  - tetap ada ketergantungan ke runtime Node/npm di VPS; bila runtime Node tidak terpasang sama sekali, deploy harus memperbaiki provisioning server dulu.

## 2026-03-15 - CI/CD Deploy Hardened with `gas build` Non-Interactive

- Scope:
  - mengganti flow deploy GitHub Actions dari build manual (`go build` + `npm run build` + restart PM2) ke jalur deploy server yang terbukti stabil: `gas build`.
  - menstandarkan deploy backend/frontend agar 100% non-interactive menggunakan `--no-ui --yes`.
  - menambahkan verifikasi runtime wajib pasca deploy: status PM2 harus `online` dan port backend/frontend harus dalam kondisi LISTEN.
  - menambahkan fail-fast script (`set -Eeuo pipefail`) + pesan error eksplisit agar kegagalan langkah terlihat jelas di log workflow.
- Dampak area:
  - `.github/workflows/deploy.yml`
  - `docs/DEPLOYMENT.md`
  - `docs/CHANGELOG_FOR_AGENTS.md`
- Command final yang dipakai:
  - backend: `gas build --no-ui --yes --type go --pm2-name jedug-backend --port 5000 --git-pull no`
  - frontend: `gas build --no-ui --yes --type node-web --pm2-name jedug-frontend --port 5001 --git-pull no`
- Mismatch baru (jika ada):
  - opsi strategy/install-deps spesifik `gas` belum dipin karena flow `--type node-web` saat ini sudah cukup untuk project frontend yang berjalan, dan menjaga kompatibilitas lintas versi `gas`.

## 2026-03-15 - Submit Merge Fix: SQLSTATE 42P08 on issue aggregate update

- Scope:
  - memperbaiki kegagalan `POST /api/v1/reports` pada jalur duplicate-merge saat update aggregate issue existing.
  - menghilangkan ambiguitas tipe SQL parameter nullable pada query update aggregates.
- Akar masalah:
  - query `UPDATE issues` menggunakan `BTRIM($4)` pada parameter `road_name` nullable (`*string` dari Go), sehingga PostgreSQL menerima `$4` sebagai `unknown` pada kondisi tertentu dan gagal infer tipe (`SQLSTATE 42P08`).
- Perbaikan:
  - menambahkan cast eksplisit pada parameter update aggregate (`$1::int`, `$2::int`, `$3::int`, `$4::text`, `$5::bigint`) agar tipe selalu deterministik.
  - menambahkan log error kontekstual baru `[REPORT] update_aggregates_query_failed` yang menyertakan `issue`, path `duplicate_merge`, serta nilai agregat input (tanpa data sensitif).
- Dampak area:
  - `backend/internal/repository/report_repository.go`
- File docs yang diupdate:
  - `docs/CHANGELOG_FOR_AGENTS.md`
- Mismatch baru (jika ada):
  - tidak ada; perubahan bersifat internal query typing dan tidak mengubah kontrak API.

## 2026-03-15 - Follow Issue API 404 contract sync fix

- Scope:
  - memperbaiki mismatch kontrak follow issue yang menyebabkan request frontend ke follow-status / count / follow berujung 404 pada runtime tertentu.
  - menambahkan alias route backend yang kompatibel dan fallback 404 terarah di helper frontend.
  - mengurangi request awal follow state agar tidak menembak endpoint ganda tanpa kebutuhan.
- Dampak area:
  - `backend/internal/http/router.go`
  - `backend/internal/http/handlers/issue_follow_handler.go`
  - `frontend/src/lib/api/issues.ts`
  - `frontend/src/routes/issues/[id]/+page.svelte`
- File docs yang diupdate:
  - `docs/BACKEND.md`
  - `docs/FRONTEND.md`
  - `docs/CHANGELOG_FOR_AGENTS.md`
- Mismatch baru (jika ada):
  - contract final tetap dipertahankan pada path utama `/follow`, `/followers/count`, `/follow-status`; alias hanya disediakan untuk kompatibilitas dan rollback-safe deploy.

## 2026-03-14 - Follow Issue / Subscribe Update MVP

- Scope:
  - menambahkan fitur follow issue anonim di detail issue publik tanpa login penuh.
  - menambahkan tabel `issue_followers` beserta endpoint follow/unfollow/count/status.
  - menambahkan card UI `Ikuti Perkembangan` di `/issues/[id]` dengan follower count, loading state, dan error copy manusiawi.
- Dampak area:
  - `backend/migrations/202603140003_create_issue_followers.sql`
  - `backend/internal/domain/issue_follow.go`
  - `backend/internal/repository/issue_follow_repository.go`
  - `backend/internal/service/issue_follow_service.go`
  - `backend/internal/service/issue_follow_service_test.go`
  - `backend/internal/http/handlers/issue_follow_handler.go`
  - `backend/internal/http/router.go`
  - `frontend/src/lib/api/client.ts`
  - `frontend/src/lib/api/issues.ts`
  - `frontend/src/lib/api/types.ts`
  - `frontend/src/lib/utils/storage.ts`
  - `frontend/src/routes/issues/[id]/+page.svelte`
- File docs yang diupdate:
  - `docs/BACKEND.md`
  - `docs/SCHEMA.md`
  - `docs/FRONTEND.md`
  - `docs/DECISIONS.md`
  - `design-docs/component-spec.md`
  - `design-docs/guide.md`
  - `docs/CHANGELOG_FOR_AGENTS.md`
- Mismatch baru (jika ada):
  - `follower_id` untuk follow issue berbeda dari `anon_token/device` submit report; ini sengaja dipisah agar MVP subscribe tetap ringan dan tidak mengikat flow bootstrap device yang sudah live.

## 2026-03-14 - Post-bugfix cleanup: trim noisy debug logs

- Scope:
  - merapikan log/debug sementara setelah fase incident debugging di area submit report, bootstrap readiness, dan location label.
- Log yang dihapus:
  - frontend `console.debug` untuk payload location label di `/lapor`.
  - frontend `console.error`/`console.warn` yang hanya mengulang error yang sudah diterjemahkan ke UI state (`bootstrap init failed`, `bootstrap ensure failed`, `submit failed`, retry token mismatch).
  - backend location-label request/success/miss/start logs yang terlalu chatty untuk setiap lookup.
  - backend report submit start / payload / per-step success logs yang berlebihan dan sempat membawa detail token/payload.
- Log yang dipertahankan:
  - backend error logs untuk parse/validation/report failure, duplicate decision, media persist failure, tx begin/commit failure.
  - backend location internal/reverse errors.
  - frontend `IssueMap` fallback logs saat heatmap/clustering gagal, karena itu sinyal degradasi fitur yang tetap berguna.
- Hasil observability:
  - log production lebih ringkas, tidak spam di path normal.
  - engineer tetap bisa mengidentifikasi failure penting tanpa payload berlebihan atau data sensitif parsial.

## 2026-03-14 - Submit Report 500 Fix: Missing submission_media table

- Akar masalah terkonfirmasi dari log produksi:
  - `ERROR: relation "submission_media" does not exist (SQLSTATE 42P01)` saat step `create_media`.
- Bukti konsistensi kode:
  - write path: `backend/internal/repository/report_repository.go` melakukan `INSERT INTO submission_media`.
  - read path: `backend/internal/repository/issue_repository.go` dan `admin_repository.go` juga query `submission_media`.
  - docs schema: `docs/SCHEMA.md` mendefinisikan `submission_media` sebagai tabel inti.
  - migrations repo: kosong, tidak ada `CREATE TABLE submission_media`.
- Perbaikan:
  1. Tambah migration `backend/migrations/202603140002_create_submission_media.sql`.
  2. Tambah sentinel error repository `ErrSubmissionMediaPersistFailed` saat insert media gagal.
  3. Mapping service error ke `ErrMediaPersist`.
  4. Handler return error terstruktur:
     - HTTP 500
     - `error_code: MEDIA_PERSIST_FAILED`
     - `message: failed to persist submission media`
- Action deploy wajib:
  - Jalankan: `psql "$DATABASE_URL" -f migrations/202603140002_create_submission_media.sql`
  - Verifikasi tidak ada lagi log `relation "submission_media" does not exist` saat submit report.

## 2026-03-14 - Submit Report 500: Structured error_code + Per-Step Logging + Migration File

- Scope: production incident debugging akhir end-to-end pada POST /api/v1/reports.
- Temuan:
  - File migrasi `backend/migrations/202603140001_create_issue_events.sql` disebutkan di changelog
    sebelumnya tetapi TIDAK ADA di repo (migrations/ kosong). File ini sekarang dibuat.
  - Response error tidak memiliki `error_code` — frontend tidak bisa memisahkan jenis error.
  - Repository tidak mem-wrap error dengan konteks, sehingga log server hanya menampilkan
    error mentah dari pgx tanpa menandai step mana yang gagal.
- Perbaikan:
  1. Membuat `backend/migrations/202603140001_create_issue_events.sql` — WAJIB DIJALANKAN DI PROD.
  2. Menambah `ErrorCode string` ke `response.Response` + helper `response.ErrorWithCode()`.
  3. Mengubah semua error path di `report_handler.go` untuk menggunakan `ErrorWithCode` dengan
     kode stabil: INVALID_PAYLOAD, PHOTO_REQUIRED, MEDIA_INVALID, LOCATION_NOT_READY,
     DEVICE_NOT_READY, DEVICE_BANNED, RATE_LIMITED, LOW_TRUST, INTERNAL_ERROR.
  4. Menambah fungsi `classifyValidationError()` di handler untuk memetakan pesan validasi
     ke error_code yang tepat tanpa mengubah kontrak `validateReportBody`.
  5. Menambah log entry/payload di awal handler + log detail di setiap titik error repository
     (tx begin, resolve region, duplicate lookup, create issue, create submission, create media,
     update aggregates, commit).
  6. Mem-wrap setiap error repository dengan `fmt.Errorf("step: %w", err)` agar log server
     menampilkan step yang gagal, misal: `error=create issue: ERROR: null value in column...`.
  7. Menambah log di service layer saat `reportRepo.SubmitReport` gagal.
- Error codes yang diperkenalkan:
  - INVALID_PAYLOAD → 400 — payload tidak bisa di-parse atau field wajib tidak valid
  - PHOTO_REQUIRED → 400 — tidak ada media dikirim
  - MEDIA_INVALID → 400 — object_key / mime_type / size_bytes media tidak valid
  - LOCATION_NOT_READY → 400 — latitude/longitude di luar range atau tidak ada
  - DEVICE_NOT_READY → 401 — anon_token tidak ada atau device belum di-bootstrap
  - DEVICE_BANNED → 403 — device di-ban
  - RATE_LIMITED → 429 — cooldown submission aktif (termasuk retry_after: 120)
  - LOW_TRUST → 403 — trust score device terlalu rendah
  - INTERNAL_ERROR → 500 — DB error atau error tak terduga lainnya
- Dampak area:
  - `backend/migrations/202603140001_create_issue_events.sql` (baru — WAJIB RUN DI PROD)
  - `backend/internal/http/response/response.go`
  - `backend/internal/http/handlers/report_handler.go`
  - `backend/internal/repository/report_repository.go`
  - `backend/internal/service/report_service.go`
- Action wajib setelah deploy:
  - Jalankan migration: `psql "$DATABASE_URL" -f migrations/202603140001_create_issue_events.sql`
  - Monitor log `[REPORT] submit_tx_committed` untuk konfirmasi submit berhasil end-to-end.
  - Monitor log `[REPORT] submit_internal_error` — jika masih ada, error_code INTERNAL_ERROR
    sekarang disertai konteks step (mis: `error=create issue: ...`) untuk diagnostik lebih cepat.

## 2026-03-14 - Submit Report 500 Fix: issue_events Migration + Non-Fatal Event Inserts

- Akar masalah:
  - Tabel `issue_events` belum pernah dibuat di database (migrations/ kosong).
  - `createIssueEvent()` dipanggil _di dalam_ transaction utama submit report.
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

- schema source of truth sekarang sudah masuk repo, tetapi environment lama tetap perlu diverifikasi apakah sudah selevel dengan baseline + migration chain repo.
- Konfigurasi PM2/Nginx runtime belum versioned di repo.
- auth admin runtime masih env + in-memory session, belum memakai tabel user/session schema.
- sebagian query backend masih defensif terhadap status issue historis `verified` / `in_progress`, sementara baseline schema issue repo tetap kanonik.

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

## 2026-03-20 - Deploy Hardening: Preflight, HTTP Readiness, dan Rollback Minimum

- Scope:
  - memperkeras workflow deploy VPS + PM2 + GitHub Actions tanpa mengganti infra dasar.
  - menambah preflight backend untuk validasi env, DB connectivity, dan init router sebelum runtime disentuh.
  - menambah readiness HTTP nyata untuk backend dan frontend, plus rollback minimum ke commit sebelumnya bila rollout gagal.
- Dampak area:
  - `.github/workflows/deploy.yml`
  - `backend/cmd/preflight/main.go`
  - `frontend/src/routes/health/+server.ts`
  - `docs/DEPLOYMENT.md`
  - `docs/FRONTEND.md`
  - `docs/CHANGELOG_FOR_AGENTS.md`
- File docs yang diupdate:
  - `docs/DEPLOYMENT.md`
  - `docs/FRONTEND.md`
  - `docs/CHANGELOG_FOR_AGENTS.md`
- Mismatch baru (jika ada):
  - rollback masih in-place pada repo kerja yang sama; mixed-version risk turun signifikan, tetapi deploy belum sepenuhnya atomic seperti model release directory/symlink.

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

## 2026-03-20 - SSE Replay Cursor + Notification Reconciliation Ringan

- Scope:
  - memperkuat reliability notification realtime tanpa mengganti SSE dan tanpa membangun sistem replay penuh.
- Dampak area:
  - `backend/internal/http/handlers/notification_handler.go`
  - `backend/internal/http/handlers/notification_handler_test.go`
  - `backend/internal/repository/notification_repository.go`
  - `backend/internal/repository/nearby_alert_repository.go`
  - `backend/internal/service/notification_service.go`
  - `backend/internal/sse/hub.go`
  - `backend/internal/sse/hub_test.go`
  - `frontend/src/lib/stores/notifications.ts`
- File docs yang diupdate:
  - `docs/BACKEND.md`
  - `docs/FRONTEND.md`
  - `docs/DEPLOYMENT.md`
  - `docs/DECISIONS.md`
  - `docs/CHANGELOG_FOR_AGENTS.md`
- Mismatch baru (jika ada):
  - replay SSE tetap ringan dan dibatasi snapshot/history terbaru; notifikasi realtime masih process-local dan belum cross-instance.

## 2026-03-20 - Deploy Path Tidak Lagi Menjalankan pm2-logrotate

- Scope:
  - mengeluarkan setup `pm2-logrotate` dari critical deploy path agar GitHub Actions tidak macet pada command PM2 module di shell non-interactive.
- Dampak area:
  - `.github/workflows/deploy.yml`
- File docs yang diupdate:
  - `docs/DEPLOYMENT.md`
  - `docs/CHANGELOG_FOR_AGENTS.md`
- Mismatch baru (jika ada):
  - `pm2-logrotate` tetap perlu dipasang manual sekali di VPS; workflow deploy tidak lagi menegakkan konfigurasi modul tersebut.

## 2026-03-20 - Sinkronisasi Label Lokasi Report ke Issue Publik

- Scope:
  - menutup gap ketika preview lokasi di `/lapor` sudah manusiawi tetapi issue publik/detail masih jatuh ke fallback koordinat atau wilayah kosong.
- Dampak area:
  - `backend/internal/service/report_location_normalizer.go`
  - `backend/internal/service/report_service.go`
  - `backend/internal/repository/report_repository.go`
  - `backend/internal/repository/issue_repository.go`
  - `backend/internal/domain/issue.go`
  - `frontend/src/lib/api/types.ts`
  - `frontend/src/lib/utils/issue-detail.ts`
  - `frontend/src/routes/issues/[id]/+page.svelte`
  - `frontend/src/routes/api/og/issues/[id]/+server.ts`
  - `frontend/src/lib/components/IssueCard.svelte`
  - `frontend/src/lib/components/IssueBottomSheet.svelte`
  - `frontend/src/routes/stats/+page.svelte`
- File docs yang diupdate:
  - `docs/BACKEND.md`
  - `docs/FRONTEND.md`
  - `docs/CHANGELOG_FOR_AGENTS.md`
- Mismatch baru (jika ada):
  - `road_type` masih tetap kosong bila geocoder/schema tidak menyediakan sumber yang tepercaya; patch ini hanya memperbaiki sinkronisasi label lokasi manusiawi, bukan klasifikasi tipe jalan.

## 2026-03-20 - Polish Presentasi Issue Detail

- Scope:
  - merapikan issue detail agar lokasi lebih manusiawi, koordinat tidak dominan, dan copy waktu/status lebih konsisten tanpa redesign total layout.
- Dampak area:
  - `frontend/src/routes/issues/[id]/+page.svelte`
  - `frontend/src/lib/components/IssueHeader.svelte`
  - `frontend/src/lib/components/IssueStats.svelte`
  - `frontend/src/lib/utils/issue-detail.ts`
  - `frontend/src/lib/utils/date.ts`
  - `design-docs/guide.md`
  - `design-docs/component-spec.md`
- File docs yang diupdate:
  - `docs/FRONTEND.md`
  - `design-docs/guide.md`
  - `design-docs/component-spec.md`
  - `docs/CHANGELOG_FOR_AGENTS.md`
- Mismatch baru (jika ada):
  - `road_type` tetap hanya muncul jika backend punya nilai yang tepercaya; UI sekarang sengaja memilih menyembunyikan field kosong daripada menampilkan placeholder noisy.

## 2026-03-20 - Persist Wilayah Administratif Dari Submit Report

- Scope:
  - memastikan field `Wilayah` di issue detail tidak jatuh ke kosong ketika submit report sudah berhasil mendapatkan label administratif dari normalisasi lokasi, terutama saat `region_id` internal tidak tersedia.
- Dampak area:
  - `backend/internal/service/reverse_geocoder.go`
  - `backend/internal/service/report_location_normalizer.go`
  - `backend/internal/service/report_service.go`
  - `backend/internal/repository/report_repository.go`
  - `backend/internal/repository/issue_repository.go`
  - `backend/migrations/202603200005_persist_submission_admin_location.sql`
  - `backend/schema/20260320_000000_baseline.sql`
  - `backend/scripts/verify_schema_governance.sh`
- File docs yang diupdate:
  - `docs/BACKEND.md`
  - `docs/SCHEMA.md`
  - `docs/CHANGELOG_FOR_AGENTS.md`
- Mismatch baru (jika ada):
  - issue lama yang historisnya tidak punya `region_id` dan belum pernah menyimpan label administratif tetap tidak bisa diperkaya mundur tanpa proses backfill/geocode ulang; patch ini menutup jalur submit baru dan issue lama yang mendapat submission baru.

## 2026-03-20 - Perluas Pemanfaatan Response Nominatim

- Scope:
  - memperkaya mapper reverse geocode Nominatim agar JEDUG memakai `accept-language=id` secara konsisten, membaca field administratif dan klasifikasi lokasi lebih lengkap, lalu mengalirkannya ke normalisasi lokasi dan statistik publik.
- Dampak area:
  - `backend/internal/service/reverse_geocoder.go`
  - `backend/internal/service/location_service.go`
  - `backend/internal/service/report_location_normalizer.go`
  - `backend/internal/repository/stats_repository.go`
  - `backend/internal/service/reverse_geocoder_test.go`
- File docs yang diupdate:
  - `docs/BACKEND.md`
  - `docs/CHANGELOG_FOR_AGENTS.md`
- Mismatch baru (jika ada):
  - metadata klasifikasi lokasi Nominatim (`category/type/addresstype/place_rank`) sudah diparse, tetapi belum dipromosikan ke field publik baru karena `road_type` existing masih sempit dan bersemantik jaringan jalan, bukan OSM place type umum.

## 2026-03-20 - Fresh DB Sekarang Benar-Benar Reset Data

- Scope:
  - menyamakan arti command fresh database dengan ekspektasi operasional: clear seluruh data schema `public` sebelum bootstrap baseline + migration.
- Dampak area:
  - `backend/scripts/bootstrap_db.sh`
  - `backend/Makefile`
  - `backend/README.md`
  - `backend/schema/README.md`
  - `backend/migrations/README.md`
- File docs yang diupdate:
  - `docs/BACKEND.md`
  - `docs/CHANGELOG_FOR_AGENTS.md`
  - `backend/README.md`
  - `backend/schema/README.md`
  - `backend/migrations/README.md`
- Mismatch baru (jika ada):
  - mode `fresh` hanya mereset schema `public`; object di schema non-`public` yang dibuat manual di database target tidak disentuh.

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
