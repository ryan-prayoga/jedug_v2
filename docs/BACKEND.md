# Backend Guide

## Stack dan Entry Point

- Bahasa: Go 1.25
- Framework HTTP: Fiber v2
- DB driver: pgx pool
- Entry point: `backend/cmd/api/main.go`
- Router: `backend/internal/http/router.go`

## Struktur Backend

- `internal/http/handlers`: endpoint handlers
- `internal/service`: business logic
- `internal/repository`: query SQL + transaksi
- `internal/storage`: local/R2 driver abstraction
- `internal/domain`: DTO/domain response model
- `internal/config`: env loading
- `schema/`: baseline SQL untuk fresh bootstrap DB
- `migrations/`: migration additive/idempotent
- `scripts/bootstrap_db.sh`: apply baseline dan/atau migrations
- `scripts/verify_schema_governance.sh`: audit schema repo vs DB aktual

## Database Bootstrap dan Governance

- Fresh DB:
  - `cd backend && DATABASE_URL=... ./scripts/bootstrap_db.sh fresh`
- Upgrade DB lama:
  - `cd backend && DATABASE_URL=... ./scripts/bootstrap_db.sh upgrade`
- Verifikasi governance schema:
  - `cd backend && DATABASE_URL=... ./scripts/verify_schema_governance.sh`
- Jika `DATABASE_URL` belum ada di shell, kedua script DB di atas akan otomatis mencoba load `backend/.env`.
- Extension yang diwajibkan oleh schema/code saat ini:
  - `postgis` untuk `GEOGRAPHY/GEOMETRY`, spatial index, `ST_*`
  - `pgcrypto` untuk `gen_random_uuid()` yang dipakai insert notifikasi/push/nearby alert

## Health + Ops Minimum

- `GET /api/v1/health` sekarang tidak hanya `db ping`, tetapi juga mengembalikan snapshot ringan:
  - `uptime_sec`
  - SSE runtime (`sse_followers`, `sse_connections`, `sse_dropped_total`)
  - push queue (`push_ready`, `push_failed_last_24h`)
  - debt retention (`notifications_over_retention`, `stale_push_subscriptions`, `disabled_push_subscriptions`, `upload_orphans_over_retention`)
  - estimasi row growth (`issue_events`, `notifications`, `push_subscriptions`, `push_delivery_jobs`, `report_upload_tickets`) dari `pg_stat_user_tables`
- backend menjalankan maintenance runner in-process dengan advisory lock DB agar hanya satu runner aktif per cluster DB:
  - interval default `6 jam`
  - bisa dijalankan manual via `cd backend && go run ./cmd/maintenance`
- retention yang berjalan otomatis:
  - hapus `notifications` lebih tua dari `90 hari`
  - soft-disable `push_subscriptions` aktif yang tidak tersentuh `180 hari`
  - hapus `push_subscriptions` disabled lebih tua dari `30 hari`
  - hapus `push_delivery_jobs.delivered` lebih tua dari `14 hari`
  - hapus `push_delivery_jobs.failed` lebih tua dari `30 hari`
  - hapus orphan upload object + ticket yang belum pernah dipakai report setelah `UPLOAD_ORPHAN_RETENTION_SEC` (default `12 jam`)
- `issue_events` sengaja belum di-prune otomatis; tabel ini masih diperlakukan sebagai histori timeline publik dan saat ini hanya di-observe pertumbuhannya lewat health snapshot.

## Route Penting

### Public API

- `GET /api/v1/health`
- `POST /api/v1/device/bootstrap`
- `POST /api/v1/device/consent`
- `POST /api/v1/uploads/presign` (`anon_token` wajib)
- `POST /api/v1/uploads/file/*` (`X-Upload-Token` wajib untuk local upload)
- `POST /api/v1/reports`
- `GET /api/v1/location/label?latitude={lat}&longitude={lng}`
- `GET /api/v1/issues`
- `GET /api/v1/issues/:id`
- `GET /api/v1/issues/:id/timeline`
- `POST /api/v1/issues/:id/follow`
- `DELETE /api/v1/issues/:id/follow`
- `GET /api/v1/issues/:id/followers/count`
- `GET /api/v1/issues/:id/follow-status?follower_id=...`
- `POST /api/v1/followers/auth`
- `GET /api/v1/nearby-alerts` (`X-Follower-Token` + `X-Device-Token` wajib)
- `POST /api/v1/nearby-alerts`
- `PATCH /api/v1/nearby-alerts/:id`
- `DELETE /api/v1/nearby-alerts/:id`
- `GET /api/v1/notification-preferences` (`X-Follower-Token` + `X-Device-Token` wajib)
- `PATCH /api/v1/notification-preferences`
- `POST /api/v1/issues/:id/flag`
- `GET /api/v1/push/status` (`X-Follower-Token` + `X-Device-Token` wajib)
- `POST /api/v1/push/subscribe`
- `POST /api/v1/push/unsubscribe`
- `GET /api/v1/stats/regions/options`
- `GET /api/v1/stats`

### Admin API

- `POST /api/v1/admin/login`
- `POST /api/v1/admin/logout`
- `GET /api/v1/admin/me` (HttpOnly session cookie; Bearer fallback masih diterima untuk kompatibilitas)
- `GET /api/v1/admin/issues`
- `GET /api/v1/admin/issues/:id`
- `POST /api/v1/admin/issues/:id/hide`
- `POST /api/v1/admin/issues/:id/unhide`
- `POST /api/v1/admin/issues/:id/fix`
- `POST /api/v1/admin/issues/:id/reject`
- `POST /api/v1/admin/devices/:id/ban`

### Admin Auth + Moderation Surface

- login admin tetap memakai env credential (`ADMIN_USERNAME`, `ADMIN_PASSWORD`), tetapi transport session kini berupa cookie `jedug_admin_session`:
  - `HttpOnly`
  - `SameSite=Strict`
  - `Path=/api/v1/admin`
  - `Secure` saat `APP_ENV=production`
- login route dilindungi dua lapis:
  - hard rate limit `10/15m` per-IP di router
  - lockout in-memory per fingerprint ringan `ip|username` setelah `5` gagal dalam window `15 menit`, lockout `30 menit`
- login failure tetap generic (`username atau password salah`) agar tidak membuka oracle username/password
- login sukses merotasi session lama untuk username yang sama; hanya satu session aktif per admin username
- `POST /api/v1/admin/logout` me-revoke session server-side dan menghapus cookie
- seluruh route `/api/v1/admin/*` sekarang mengirim header `Cache-Control: no-store`

## Flow Service/Repository Penting

### Device Bootstrap + Consent

- Service `DeviceService`:
  - hash `anon_token` (SHA-256)
  - lookup existing device by hash
  - jika tidak ada, generate token baru (raw token hanya dikembalikan ke client)
- Repository `DeviceRepository`:
  - simpan `anon_token_hash`, IP, user-agent
  - catat consent ke `device_consents`

### Report Submission

- Upload hardening sebelum submit:
  - `POST /api/v1/uploads/presign` sekarang wajib `anon_token` yang valid.
  - backend menerbitkan `upload_token` bertanda tangan server untuk satu `object_key`, `mime_type`, `size_bytes`, dan `device_id` tertentu dengan TTL pendek (`UPLOAD_TICKET_TTL_SEC`, default 10 menit).
  - setiap presign juga mencatat row pending di `report_upload_tickets`; satu device hanya boleh punya jumlah upload belum dipakai yang terbatas dalam window pendek (`UPLOAD_PENDING_LIMIT`, default `4` dalam `UPLOAD_PENDING_WINDOW_SEC`, default `1800` detik).
  - local upload ke `POST /api/v1/uploads/file/{object_key}` sekarang wajib header `X-Upload-Token`; upload tanpa proof ini ditolak.
  - local upload sekarang juga dilindungi limiter keras `10/15m` per-IP agar binary upload tidak bisa di-churn bebas.
  - R2 tetap memakai presigned `PUT`, tetapi ownership tetap diverifikasi saat `/reports` lewat `upload_token` yang sama.
- Handler validasi payload report + media.
- Service `ReportService` enforce:
  - device harus ada
  - `client_request_id` divalidasi sebagai UUID dan diperlakukan sebagai idempotency key utama
  - replay dengan `(device_id, client_request_id)` yang sama dicek sebelum ban/trust/cooldown/upload validation, sehingga retry request yang sama mengembalikan hasil lama dan tidak tersandung cooldown
  - submission baru tetap harus lolos guard device banned + trust score minimal
  - cooldown submit 2 menit/device hanya berlaku untuk submission baru
  - payload reuse dengan `client_request_id` yang sama tetapi fingerprint request berbeda ditolak `409 IDEMPOTENCY_CONFLICT`
  - setiap `media[]` wajib membawa `upload_token`
  - `upload_token` harus cocok dengan `device_id`, `object_key`, `mime_type`, dan `size_bytes` dari media yang disubmit
  - `report_upload_tickets` untuk `object_key` tersebut harus masih ada dan cocok; ticket yang sudah dipruning/tidak dikenal ditolak
  - object storage diverifikasi benar-benar memiliki file untuk `object_key` tersebut sebelum report diterima
  - `object_key` yang sudah pernah masuk `submission_media` ditolak agar media tidak bisa dipakai ulang lintas report
  - submit report yang sukses menghapus row pending `report_upload_tickets` di transaction yang sama, sehingga cleanup orphan hanya menyentuh media yang memang belum pernah ter-link ke report
  - normalisasi lokasi sekali per laporan:
    - lookup region internal (`regions`) sebagai sumber utama label wilayah
    - reverse geocoding ringan untuk melengkapi `road_name` jika kosong
    - jika nama jalan tidak ada tetapi label area/lokalitas manusiawi tersedia, `road_name` issue kini memakai label area itu lebih dulu, bukan langsung fallback koordinat
    - fallback `Kawasan sekitar lat,lng` jika label manusiawi tidak tersedia
  - reverse geocoding memakai timeout + cache in-memory agar ringan, dan gagal geocode tidak memblok submit report
- Repository `ReportRepository` (transactional):
  - ambil advisory transaction lock ringan berbasis `(device_id, client_request_id)` untuk men-serialize retry paralel dari request yang sama tanpa menambah sistem idempotency terpisah
  - re-check existing submission di dalam transaction; jika sudah ada:
    - fingerprint sama -> return result existing
    - fingerprint beda -> reject conflict
  - `region_id` issue/submission baru kini lebih dulu memakai hasil lookup lokasi yang sudah dinormalisasi saat submit; jika kosong baru fallback ke resolver region internal repository
  - prioritas level wilayah kini juga mengenali alias Indonesia (`provinsi`, `kabupaten`, `kota`, `kecamatan`) agar `region_id` issue/submission baru tidak jatuh ke level yang terlalu bawah hanya karena label level berbeda
  - duplicate detection issue aktif publik (`open|verified|in_progress`, `is_hidden=false`) dalam radius configurable (default 30m)
  - pilih kandidat terbaik: distance terdekat -> status aktif -> verification status -> `last_seen_at` terbaru -> severity tertinggi
  - create issue baru jika tidak ada kandidat relevan
  - create `issue_submissions` dengan `request_fingerprint` + `created_issue`
  - create `submission_media`
  - saat create issue baru:
    - isi `road_name` dari hasil normalisasi lokasi
    - isi `region_id` dari lookup internal bila tersedia
  - saat merge update aggregate issue:
    - `last_seen_at = NOW()`
    - `submission_count += 1`
    - `photo_count += jumlah media submission`
    - `casualty_count = GREATEST(existing, incoming)` (hindari overcount duplicate)
    - `severity_current = GREATEST(existing, incoming)`
    - `severity_max = GREATEST(existing, incoming)`
    - isi `road_name` bila issue existing masih kosong (tanpa overwrite nilai valid)
    - isi `region_id` bila issue existing masih `NULL` (tanpa overwrite nilai valid)

### Issue Listing + Detail

- `IssueRepository.List`:
  - filter `is_hidden = false`
  - exclude status `rejected`, `merged`
  - optional `status`, `severity >=`, `bbox`
  - enrich lokasi publik bukan lagi hanya `join regions` langsung; query kini menurunkan `district_name`, `regency_name`, `province_name` dari `issues.region_id`, `issue_submissions.region_id` terbaru, atau spatial fallback dari `public_location`
  - `region_name` publik kini berisi label administratif manusiawi (`district/regency/province`) bila tersedia
- `IssueRepository.FindByID`:
  - hanya mengembalikan issue publik (`is_hidden = false`)
  - exclude status `rejected`, `merged` agar deep-link publik konsisten dengan list/map
- `FindByIDWithDetail`:
  - media publik top 20, primary first, exclude submission berstatus `rejected`
  - expose `primary_media` additive untuk hero/OG fallback, tetap kompatibel dengan array `media`
  - expose `public_note` additive sebagai ringkasan catatan publik yang sudah dinormalisasi/truncate
  - recent submissions top 3, exclude submission berstatus `rejected`
  - recent submissions membawa `casualty_count` dan `public_note` additive agar UI tidak perlu menampilkan note mentah
  - resolve `public_url` media via storage service (compatible local legacy + R2)
  - hanya expose field publik (tanpa device/admin/internal note)
  - response additive publik kini juga membawa `district_name`, `regency_name`, dan `province_name` untuk presenter/UI yang perlu fallback lokasi lebih manusiawi
- `ListTimeline`:
  - endpoint `GET /api/v1/issues/:id/timeline`
  - urut event terbaru di atas (`created_at DESC, id DESC`)
  - default/maximum `limit` dijaga di 100 untuk menjaga payload tetap ringan
  - support pagination lewat query `limit` + `offset`

### Follow Issue / Subscribe Update

- Tabel follower anonim: `issue_followers`.
- Identity follower memakai `follower_id` UUID dari browser (tanpa login).
- Hardening tambahan:
  - backend mengikat `follower_id` ke `X-Device-Token` anonim lewat tabel `follower_auth_bindings`
  - response `follow` / `unfollow` / `follow-status` kini juga dapat mengembalikan `follower_token` device-bound untuk endpoint non-SSE, plus `follower_stream_token` khusus SSE
  - endpoint `POST /api/v1/followers/auth` dipakai frontend untuk refresh token notifikasi secara aman tanpa login penuh
- Endpoint publik additive:
  - `POST /api/v1/issues/:id/follow`
  - `DELETE /api/v1/issues/:id/follow`
  - `GET /api/v1/issues/:id/followers/count`
  - `GET /api/v1/issues/:id/follow-status?follower_id=...`
  - `POST /api/v1/followers/auth`
- Alias kompatibilitas yang juga dilayani agar caller lama/tidak sinkron tidak jatuh ke 404:
  - `POST /api/v1/issues/:id/followers`
  - `DELETE /api/v1/issues/:id/followers`
  - `GET /api/v1/issues/:id/count`
  - `GET /api/v1/issues/:id/follow/status?follower_id=...`
- Handler memvalidasi:
  - `issue_id` harus UUID valid
  - `follower_id` harus UUID valid dan non-nil
  - `DELETE` / `POST` menerima `follower_id` dari body, dan fallback query param untuk kompatibilitas client/proxy yang tidak mengirim body DELETE dengan stabil
  - `follow` / `unfollow` / `follow-status` sekarang juga memerlukan `X-Device-Token` agar `follower_id` tidak lagi menjadi bearer secret mentah
- Service memastikan issue target masih issue publik (`FindByID`), sehingga hidden/rejected/merged tidak bisa di-follow dari endpoint publik.
- Repository menggunakan:
  - `INSERT ... ON CONFLICT (issue_id, follower_id) DO NOTHING` agar follow idempotent dan conflict-safe
  - `DELETE ... WHERE issue_id = $1 AND follower_id = $2` agar unfollow aman walau record belum ada
  - `COUNT(*)` per issue agar follower count akurat tanpa cache/denormalisasi dulu
- Rate limit tambahan:
  - follow/unfollow: `30/min` per-IP
- Response tetap konsisten memakai wrapper `response.OK(...)` dengan payload utama:
  - `following`
  - `followers_count`

### Issue Timeline Event Logging

- Tabel event: `issue_events`.
- Tabel notifikasi: `notifications` — di-populate otomatis oleh `DispatchNotificationsForEvent` setiap kali event berhasil diinsert.
- Dispatch function: `repository.DispatchNotificationsForEvent(ctx, db, pushNotifier, issueID, eventID, eventType, excludeFollowerID)` — free function di `notification_repository.go`, dipakai oleh `report_repository` dan `admin_repository`.
- Endpoint notifikasi in-app publik:
  - `GET /api/v1/notifications?limit=50` dengan header `X-Follower-Token` + `X-Device-Token`
  - `PATCH /api/v1/notifications/:id/read` dengan header `X-Follower-Token` + `X-Device-Token`
  - `DELETE /api/v1/notifications/:id` dengan header `X-Follower-Token` + `X-Device-Token`
  - `GET /api/v1/notifications/stream?stream_token=...&last_event_id=...` — **SSE stream** (text/event-stream) dengan replay ringan sejak `event_id` terakhir yang diketahui client
- Endpoint preferensi notifikasi publik:
  - `GET /api/v1/notification-preferences` dengan header `X-Follower-Token` + `X-Device-Token`
  - `PATCH /api/v1/notification-preferences`
- Endpoint nearby alerts publik:
  - `GET /api/v1/nearby-alerts` dengan header `X-Follower-Token` + `X-Device-Token`
  - `POST /api/v1/nearby-alerts`
  - `PATCH /api/v1/nearby-alerts/:id`
  - `DELETE /api/v1/nearby-alerts/:id`
- Endpoint browser push publik:
  - `GET /api/v1/push/status` dengan header `X-Follower-Token` + `X-Device-Token`
  - `POST /api/v1/push/subscribe`
  - `POST /api/v1/push/unsubscribe`
- `POST /api/v1/followers/auth` kini menerbitkan dua token bertanda tangan server:
  - `follower_token` untuk endpoint non-SSE, wajib diverifikasi bersama `X-Device-Token`
  - `stream_token` untuk SSE, purpose-limited dan TTL pendek agar query string tidak membawa token umum
- Semua endpoint notifikasi/push sekarang mengekstrak `follower_id` dari token bertanda tangan server; ownership tidak lagi bergantung pada UUID mentah dari caller.
- Endpoint non-SSE menolak token tanpa `X-Device-Token` yang cocok, sehingga token yang bocor tidak cukup dipakai sendiri.
- Endpoint preferensi juga memakai boundary yang sama; backend tidak menerima `follower_id` mentah sebagai bearer secret untuk mengubah settings.
- Mark-as-read tetap dikunci di DB oleh pasangan `notification_id + follower_id` dan mengembalikan `read_at` persisten; jika row tidak ditemukan, response `404`.
- Delete notification tetap dikunci oleh pasangan `notification_id + follower_id`; response `deleted: true|false` dibuat aman/idempotent tanpa membocorkan ownership follower lain.
- Storage preferensi notifikasi:
  - tabel `notification_preferences`
  - `GET /notification-preferences` dapat mengembalikan default sintetis tanpa write DB; row baru dimaterialkan saat user pertama kali menyimpan preference
  - default:
    - `notifications_enabled = true`
    - `in_app_enabled = true`
    - `push_enabled = true` hanya jika follower sudah punya subscription push aktif ketika row pertama dibuat
    - seluruh event preference default `true`, termasuk `notify_on_nearby_issue_created`
- Integrasi pipeline paling aman tetap di `DispatchNotificationsForEvent(...)`:
  - dispatcher sekarang mengevaluasi preference follower sekali untuk setiap event issue
  - row tabel `notifications` hanya dibuat untuk follower yang lolos `notifications_enabled + event_pref + in_app_enabled`
  - SSE hanya dikirim untuk row in-app yang benar-benar baru terbuat
  - browser push diantrekan terpisah untuk follower yang lolos `notifications_enabled + event_pref + push_enabled`
  - hasilnya: `in_app_enabled = false` tidak lagi ikut mematikan browser push, dan `push_enabled = false` tidak mencegah notifikasi in-app
- Self-notify prevention:
  - endpoint submit report menerima field opsional `actor_follower_id`.
  - dispatcher skip follower yang sama dengan actor (`excludeFollowerID`) agar pengirim update tidak menerima notifikasi untuk event yang ia buat sendiri.
  - behavior ini hanya diterapkan pada event dari flow submit report; event admin tetap broadcast ke seluruh follower.
- **Nearby Alerts** (`nearby_alert_subscriptions` + `nearby_alert_deliveries`):
  - memakai identity yang sama dengan notification center: `follower_id` + token follower device-bound untuk non-SSE + stream token khusus SSE.
  - subscription minimum menyimpan `latitude`, `longitude`, `radius_m`, `label`, `enabled`.
  - guard service:
    - maksimum `10` watched locations per follower/browser
    - radius valid `100..5000m`
    - latitude/longitude wajib valid, dan patch koordinat harus dikirim berpasangan
  - dispatch flow saat `issue_created` berhasil ditulis ke `issue_events`:
    1. cari subscription `enabled` yang memenuhi `ST_DWithin(issue.public_location, subscription_point, radius_m)`
    2. insert dedupe row ke `nearby_alert_deliveries` dengan unique `(subscription_id, issue_id)`
    3. group hasil per follower agar satu follower hanya menerima satu notif untuk satu issue baru meski beberapa lokasi pantauan overlap
    4. buat notification type `nearby_issue_created` + push payload bila preference/channel mengizinkan
  - `nearby_alert_deliveries` sengaja tetap diisi walau preference/channel saat itu off, agar issue lama tidak dikirim retroaktif ketika user menyalakan setting lagi.
  - self-notify juga di-skip untuk `issue_created` jika `actor_follower_id` sama dengan follower pemilik nearby alert.
- Copy notifikasi sekarang kontekstual lokasi issue:
  - prioritas label: `issues.road_name` → `regions.name` dari issue → `regions.name` dari submission terbaru → fallback `Issue #<short-id>`.
  - contoh: `Foto baru ditambahkan pada laporan di Jalan ...`.
- **SSE Hub** (`internal/sse/hub.go`):
  - Singleton `sse.Default` dipakai oleh dispatcher dan endpoint stream.
  - endpoint stream hanya menerima `stream_token` purpose `notification_stream`; token non-SSE tidak valid di jalur ini.
  - `DispatchNotificationsForEvent` kini memakai `RETURNING` untuk mendapat follower IDs yang baru di-insert, lalu memanggil `sse.Default.Push(followerID, msg)` untuk setiap row.
  - Setiap SSE connection di-buffer (channel 16 slot, non-blocking drop).
  - drop akibat buffer penuh kini dihitung kumulatif (`sse_dropped_total`) agar loss tidak lagi sepenuhnya diam-diam.
  - Koneksi di-cleanup otomatis saat client disconnect (Flush error) via `defer done()`.
  - Ping/heartbeat dikirim setiap 30 detik untuk menjaga koneksi dan mendeteksi putus.
  - Format event sekarang menyertakan `id: <event_id>` di frame SSE, lalu payload `event: notification\ndata: {id,issue_id,event_id,type,title,message,created_at}\n\n`
  - saat client reconnect dengan `last_event_id`, backend me-replay notification follower dengan `event_id > last_event_id` (limit ringan, urut ascending) sebelum stream live lanjut normal
  - handler SSE sekarang menulis log `stream_open` / `stream_close` dengan follower, replay count, durasi, dan close reason untuk diagnosis runtime.
- **Web Push notifier** (`internal/push/notifier.go`):
  - browser push tetap additive di atas tabel `notifications` + SSE.
  - dispatcher mengumpulkan row notifikasi baru lalu mengantrekan batch ke tabel outbox `push_delivery_jobs`; request submit/moderation tetap cepat karena tidak menunggu delivery ke provider.
  - worker in-process meng-claim job dari DB dengan `FOR UPDATE SKIP LOCKED`, sehingga restart/crash tidak menghilangkan batch yang belum selesai.
  - payload push minimal: `title`, `body`, `issue_id`, `url`, `type`.
  - `url` dibentuk dari `WEB_PUSH_SITE_URL + /issues/{issue_id}` agar klik notifikasi selalu menuju issue publik yang benar.
  - retry ringan dibatasi `5` attempt dengan backoff bertahap (`30s -> 2m -> 5m -> 15m`) dan status job dicatat lewat `attempt_count`, `next_attempt_at`, `delivered_at`, `failed_at`, `last_error`.
  - queue pressure tidak lagi hilang diam-diam saat channel in-memory penuh; kegagalan enqueue/delivery sekarang bisa diaudit langsung dari tabel outbox.
  - response `404/410` dari push service dianggap subscription invalid/expired dan row ditandai `disabled_at`.
  - subscribe memvalidasi endpoint Web Push hanya untuk host/path provider yang dikenal (`fcm.googleapis.com`, `updates.push.services.mozilla.com`, `push.services.mozilla.com`, `web.push.apple.com`), wajib `https`, tanpa credential, dan kunci `p256dh/auth` harus valid.
- Event dibuat otomatis saat:
  - issue baru dibuat (`issue_created`)
  - submission membawa foto (`photo_added`)
  - severity issue naik (`severity_changed`)
  - korban dilaporkan/naik (`casualty_reported`)
  - status issue berubah via moderasi (`status_updated`)
- `event_data` disimpan sebagai JSONB agar additive/fleksibel tanpa mematahkan schema timeline.

### Browser Push Subscription Storage

- Tabel `push_subscriptions` menyimpan endpoint Web Push aktif per `follower_id`.
- Endpoint subscribe menerima:
  - `follower_token`
  - `subscription.endpoint`
  - `subscription.keys.p256dh`
  - `subscription.keys.auth`
- Subscribe memakai upsert berbasis `endpoint`:
  - saat browser merotasi endpoint, row yang sama dihidupkan kembali (`disabled_at = NULL`) dan metadata diperbarui.
- Unsubscribe tidak menghapus row secara keras:
  - row ditandai `disabled_at` agar invalidation dari provider tetap aman dan historinya masih bisa diaudit.
- Status endpoint mengembalikan:
  - `enabled`
  - `subscribed`
  - `subscription_count`
  - `vapid_public_key`
- Konfigurasi:
  - jika seluruh env Web Push kosong, backend tetap hidup dan fitur browser push dianggap disabled.
  - jika env Web Push hanya terisi sebagian, startup gagal cepat untuk mencegah half-config di production.
  - `UPLOAD_TOKEN_SECRET` dipakai untuk menandatangani `upload_token`; jika kosong backend fallback ke `FOLLOWER_TOKEN_SECRET`.
  - `UPLOAD_TICKET_TTL_SEC` mengatur TTL upload ticket (default `600` detik).
  - `UPLOAD_PENDING_WINDOW_SEC` default `1800` detik untuk menghitung backlog upload belum dipakai per device.
  - `UPLOAD_PENDING_LIMIT` default `4`; jika terlampaui, presign baru ditolak `429`.
  - `FOLLOWER_TOKEN_SECRET` wajib ada dan minimal 32 karakter; backend memakai secret ini untuk menandatangani `follower_token`.
  - `FOLLOWER_TOKEN_TTL_SEC` default sekarang `43200` detik (12 jam) untuk token non-SSE.
  - `FOLLOWER_STREAM_TOKEN_TTL_SEC` default `600` detik (10 menit) untuk token SSE/query string.
  - env ops/retention:
    - `MAINTENANCE_ENABLED` default `true`
    - `MAINTENANCE_INTERVAL_SEC` default `21600`
    - `NOTIFICATIONS_RETENTION_DAYS` default `90`
    - `PUSH_SUBSCRIPTIONS_STALE_DAYS` default `180`
    - `PUSH_SUBSCRIPTIONS_DISABLED_RETENTION_DAYS` default `30`
    - `PUSH_DELIVERY_DELIVERED_RETENTION_DAYS` default `14`
    - `PUSH_DELIVERY_FAILED_RETENTION_DAYS` default `30`
    - `UPLOAD_ORPHAN_RETENTION_SEC` default `43200` (12 jam)

### Log Signals Minimum

- request/runtime log tetap lewat middleware Fiber + PM2 logs.
- tag log yang sekarang cukup penting untuk diagnosis:
  - `[REPORT]` untuk parse/idempotency/internal failure submit report
  - `[ANTISPAM]` untuk rate limit/cooldown/ban/low-trust
  - `[ADMIN]` untuk login, moderation, ban, serta audit/event moderation yang gagal
  - `[SSE]` untuk stream open/close dan alasan putus
  - `[PUSH]` untuk enqueue/claim/send/retry/fail browser push
  - `[OPS]` untuk hasil maintenance retention periodik

### Public Stats Dashboard (`GET /api/v1/stats`)

- Endpoint menerima query opsional:
  - `province_id`
  - `regency_id`
- Endpoint opsi wilayah terpisah:
  - `GET /api/v1/stats/regions/options`
  - payload mengembalikan daftar `provinces[]` dengan `regencies[]` per provinsi agar frontend bisa merender dropdown manual tanpa bergantung pada snapshot `/stats` saat itu
- `StatsService` memakai cache in-memory thread-safe (`sync.RWMutex`) dengan TTL 45 detik.
- Cache stats dibedakan per kombinasi `province_id + regency_id` agar scope wilayah tidak saling tertukar.
- Cache terpisah juga dipakai untuk endpoint opsi wilayah agar dropdown publik tetap ringan.
- Jika query DB gagal sesaat, service fallback ke cache stale agar endpoint tetap responsif.
- `StatsRepository` menjalankan query agregasi ringan berbasis tabel `issues` + join hirarki `regions`:
  - sumber region issue stats sekarang memakai prioritas:
    1. `issues.region_id`
    2. `latest issue_submissions.region_id`
    3. spatial fallback dari `issues.public_location`
  - normalisasi level wilayah mengenali bentuk English + Indonesia (`province/provinsi`, `city/kota`, `regency/kabupaten`, `district/kecamatan`) dan menelusuri ancestor sampai level provinsi walau `base_region_id` issue lama berada di level desa/kelurahan
  - snapshot global (`global`):
    - `total_issues`
    - `total_issues_this_week`
    - `total_casualties`
    - `total_photos`
    - `total_reports`
  - summary scope aktif (`summary`):
    - memakai scope hasil filter aktif (`province_id` / `regency_id`) atau fallback default backend bila query kosong
    - field totals sama dengan snapshot global, tetapi **bukan** campuran global-vs-scoped lagi
  - scope metadata:
    - `active_scope.kind`: `global | province | regency`
    - `active_scope.label`: label manusiawi untuk scope aktif
    - `active_scope.is_default`: `true` bila scope aktif berasal dari fallback default backend, bukan query eksplisit caller
  - status stats (`status`) sekarang mengikuti scope aktif yang sama dengan `summary`
    - `open` dihitung sebagai issue unresolved (`open|verified|in_progress`)
    - `fixed`
    - `archived`
  - time stats (`time`) sekarang mengikuti scope aktif yang sama dengan `summary`
    - `average_issue_age_days`
    - `oldest_open_issue_age_days`
    - metadata issue tertua unresolved (`oldest_open_issue_id`, lokasi, first seen) jika ada
  - filter metadata:
    - `filters.province_options`
    - `filters.regency_options` untuk provinsi aktif
    - `filters.active_province_id`, `filters.active_regency_id`
    - `filters.scope_label`
    - jika query filter kosong, backend fallback ke provinsi + kabupaten/kota teratas agar frontend selalu punya default yang masuk akal
    - jika caller hanya mengirim `province_id`, backend mempertahankan scope level provinsi dan tidak lagi memaksa memilih kabupaten/kota pertama
  - region leaderboard:
    - daftar wilayah administratif di scope aktif dengan identity stabil berbasis `regions.id`
    - prioritas grouping: `district_id` -> `regency_id` -> `province_id`
    - row tanpa identity administratif stabil tidak ikut leaderboard agar tidak tercampur oleh `GROUP BY name`
    - ranking berdasarkan `issue_count`, lalu `report_count`, lalu `casualty_count`
    - payload leaderboard sekarang additive membawa:
      - `region_id`
      - `region_level`
      - `region_name`
      - `parent_region_name`
      - `regency_name`
      - `province_name`
  - top issues:
    - issue dengan laporan terbanyak di scope wilayah aktif
    - issue dengan korban terbanyak di scope wilayah aktif
    - issue paling lama belum diperbaiki di scope wilayah aktif
    - payload issue membawa `district_name`, `regency_name`, `province_name`, serta `region_name` administratif fallback
    - `region_name` tidak lagi jatuh ke copy `Sekitar Jalan ...`; title issue tetap memakai `road_name` terpisah
- Semua query stats hanya memakai data publik issue:
  - `is_hidden = false`
  - status `rejected` dan `merged` dikecualikan
- Current implementation:
  - `global` tetap tersedia sebagai snapshot seluruh data publik untuk pembanding, tetapi summary cards frontend kini membaca `summary` yang sudah scoped.
- Known mismatch:
  - issue yang belum bisa di-resolve ke identity administratif stabil tetap masuk summary/top issue scoped, tetapi sengaja tidak ikut leaderboard agar ranking wilayah tidak misleading.
- Handler menambah header HTTP cache:
  - `Cache-Control: public, max-age=30, stale-while-revalidate=30`

### Backfill Lokasi Issue Lama (helper opsional)

- Tersedia command helper:
  - `go run ./cmd/backfill_issue_location --limit=200`
- Tujuan:
  - mengisi `road_name` issue lama yang kosong
  - mengisi `region_id` issue lama yang masih `NULL`
- Command menggunakan normalizer yang sama dengan flow submit report:
  - lookup region internal + reverse geocoding ringan + fallback koordinat
- Mode aman:
  - `--dry-run` untuk preview tanpa update data.

### Location Label Resolve (UX `/lapor`)

- `LocationHandler.ResolveLabel` menerima query `latitude` + `longitude`.
- `LocationService` memanggil repository untuk lookup wilayah internal dari tabel `regions`, lalu fallback ke reverse geocoding jika tidak ada match.
- Repository memilih polygon wilayah terkecil yang menutupi titik (`ST_Covers` + `ORDER BY ST_Area ASC LIMIT 1`) agar label lebih manusiawi.
- Response selalu aman untuk UX:
  - jika internal region ketemu: kirim `label`, `region_name`, `region_level`, parent chain, serta additive `district_name`, `regency_name`, `province_name`, `source=internal_regions`.
  - jika internal region miss tapi reverse geocode berhasil: kirim label fallback manusiawi (`source=reverse_geocode`).
  - jika semua sumber miss: field label bernilai `null`, `source=unresolved`, tanpa memblok submit report.
- `region_level` kini dinormalisasi ke bentuk kanonik (`province`, `city`, `regency`, `district`, `subdistrict`, `village`) meski data master memakai alias Indonesia.
- Debug log pipeline location label:
  - request masuk + koordinat
  - hasil query internal (`hit/miss/error`)
  - status reverse geocode (`start/hit/error/empty`)
  - response final yang dikirim handler
- Endpoint ini **hanya** untuk konfirmasi UX lokasi, bukan pengganti source of truth geospatial issue/submission.

### Moderation

- `AdminService`:
  - login credential dari env
  - session token in-memory TTL 24 jam
  - action hide/unhide/fix/reject/ban
  - target moderation yang tidak ada sekarang diperlakukan eksplisit sebagai `404`, bukan success kosong
  - hide/unhide/ban mengecek `RowsAffected` agar tidak ada false success
  - fix/reject memakai satu transaksi domain untuk:
    - lock issue target
    - validasi target ada
    - update `issues.status`
    - adjust trust score submitter hanya jika status benar-benar berubah
  - `issue_events` (`status_updated`) dan `moderation_actions` sekarang dijalankan post-commit sebagai best-effort audit
  - kegagalan audit/event hanya di-log dan tidak membatalkan action utama yang sudah committed, agar operator tidak menerima `500` palsu setelah perubahan domain sebenarnya berhasil
  - repeated fix/reject ke status yang sama tidak lagi mengulang trust adjustment atau memproduksi `status_updated` event tambahan

### Community Flag

- `FlagService`:
  - validasi reason flag
  - dedup per `(issue_id, device_id)`
  - auto-hide issue jika unique flag >= 3
  - log auto-hide sebagai action `system`

## Middleware dan Hardening

- CORS + logger + recover + request-id
- rate limiter per-IP:
  - bootstrap 10/min
  - consent 10/min
  - presign 10/15m
  - upload file 10/15m
  - report 5/min
  - issue flag 10/min

## Kontrak Response Sensitif

`Issue` response shape dipakai lintas:

- Map marker rendering
- Bottom sheet
- Public detail page
- Admin issue list/detail

Perubahan field berikut berisiko tinggi:

- `latitude`, `longitude`
- `status`
- `severity_current`
- `submission_count`, `photo_count`, `casualty_count`
- `flag_count`
- `region_name` (field turunan dari tabel `regions` yang dipakai detail page publik)

Untuk `GET /api/v1/issues/:id`, field additive publik yang dipakai halaman detail production-ready saat ini:

- `primary_media`
- `public_note`
- `recent_submissions[].public_note`
- `recent_submissions[].casualty_count`

Backward compatibility dijaga dengan tidak menghapus field lama seperti `media`, `recent_submissions[].note`, atau shape `Issue` yang dipakai map/list/admin.

## Current Implementation

- Pattern layering rapi dan konsisten.
- Storage sudah driver-based (`local`/`r2`) dengan fallback media legacy local.
- Moderation dan trust logic sudah aktif di jalur utama.

## Known Mismatch

- `issue_status_history` belum diisi otomatis saat status berubah.
- Field lifecycle seperti `resolved_at` belum di-maintain service.
- Auth admin belum memakai `users/user_sessions` database.

## Read This Next

- `docs/SCHEMA.md`
- `docs/STORAGE_AND_MEDIA.md`
- `docs/MODERATION.md`
- `docs/DEPLOYMENT.md`
