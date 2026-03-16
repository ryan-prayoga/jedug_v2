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
