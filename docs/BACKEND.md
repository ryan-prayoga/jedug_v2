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

## Route Penting

### Public API

- `GET /api/v1/health`
- `POST /api/v1/device/bootstrap`
- `POST /api/v1/device/consent`
- `POST /api/v1/uploads/presign`
- `POST /api/v1/uploads/file/*`
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
- `GET /api/v1/nearby-alerts?follower_token=...`
- `POST /api/v1/nearby-alerts`
- `PATCH /api/v1/nearby-alerts/:id`
- `DELETE /api/v1/nearby-alerts/:id`
- `GET /api/v1/notification-preferences?follower_token=...`
- `PATCH /api/v1/notification-preferences`
- `POST /api/v1/issues/:id/flag`
- `GET /api/v1/push/status?follower_token=...`
- `POST /api/v1/push/subscribe`
- `POST /api/v1/push/unsubscribe`
- `GET /api/v1/stats`

### Admin API

- `POST /api/v1/admin/login`
- `GET /api/v1/admin/me` (Bearer token)
- `GET /api/v1/admin/issues`
- `GET /api/v1/admin/issues/:id`
- `POST /api/v1/admin/issues/:id/hide`
- `POST /api/v1/admin/issues/:id/unhide`
- `POST /api/v1/admin/issues/:id/fix`
- `POST /api/v1/admin/issues/:id/reject`
- `POST /api/v1/admin/devices/:id/ban`

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

- Handler validasi payload report + media.
- Service `ReportService` enforce:
  - device harus ada, tidak banned
  - trust score minimal
  - cooldown submit 2 menit/device
  - idempotency via `client_request_id`
  - normalisasi lokasi sekali per laporan:
    - lookup region internal (`regions`) sebagai sumber utama label wilayah
    - reverse geocoding ringan untuk melengkapi `road_name` jika kosong
    - fallback `Kawasan sekitar lat,lng` jika label manusiawi tidak tersedia
  - reverse geocoding memakai timeout + cache in-memory agar ringan, dan gagal geocode tidak memblok submit report
- Repository `ReportRepository` (transactional):
  - resolve region internal terbaik (prioritas district, fallback smallest covering region)
  - duplicate detection issue aktif publik (`open|verified|in_progress`, `is_hidden=false`) dalam radius configurable (default 30m)
  - pilih kandidat terbaik: distance terdekat -> status aktif -> verification status -> `last_seen_at` terbaru -> severity tertinggi
  - create issue baru jika tidak ada kandidat relevan
  - create `issue_submissions`
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
  - enrich lokasi dengan `region_name` (join `regions`) untuk kebutuhan UI publik
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
  - response `follow` / `unfollow` / `follow-status` kini juga dapat mengembalikan `follower_token` bertanda tangan server + `follower_token_expires_at`
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
  - `GET /api/v1/notifications?follower_token=...&limit=50`
  - `PATCH /api/v1/notifications/:id/read?follower_token=...`
  - `DELETE /api/v1/notifications/:id?follower_token=...`
  - `GET /api/v1/notifications/stream?follower_token=...` — **SSE stream** (text/event-stream)
- Endpoint preferensi notifikasi publik:
  - `GET /api/v1/notification-preferences?follower_token=...`
  - `PATCH /api/v1/notification-preferences`
- Endpoint nearby alerts publik:
  - `GET /api/v1/nearby-alerts?follower_token=...`
  - `POST /api/v1/nearby-alerts`
  - `PATCH /api/v1/nearby-alerts/:id`
  - `DELETE /api/v1/nearby-alerts/:id`
- Endpoint browser push publik:
  - `GET /api/v1/push/status?follower_token=...`
  - `POST /api/v1/push/subscribe`
  - `POST /api/v1/push/unsubscribe`
- Semua endpoint notifikasi/push sekarang mengekstrak `follower_id` dari `follower_token` bertanda tangan server; ownership tidak lagi bergantung pada UUID mentah dari caller.
- Endpoint preferensi juga memakai `follower_token` yang sama; backend tidak menerima `follower_id` mentah sebagai bearer secret untuk mengubah settings.
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
  - memakai identity yang sama dengan notification center: `follower_id` + `follower_token` + device-bound follower auth.
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
  - `DispatchNotificationsForEvent` kini memakai `RETURNING` untuk mendapat follower IDs yang baru di-insert, lalu memanggil `sse.Default.Push(followerID, msg)` untuk setiap row.
  - Setiap SSE connection di-buffer (channel 16 slot, non-blocking drop).
  - Koneksi di-cleanup otomatis saat client disconnect (Flush error) via `defer done()`.
  - Ping/heartbeat dikirim setiap 30 detik untuk menjaga koneksi dan mendeteksi putus.
  - Format event: `event: notification\ndata: {id,issue_id,event_id,type,title,message,created_at}\n\n`
- **Web Push notifier** (`internal/push/notifier.go`):
  - browser push tetap additive di atas tabel `notifications` + SSE.
  - dispatcher mengumpulkan row notifikasi baru lalu mengantrekan batch ke worker in-process; request submit/moderation tidak lagi menunggu seluruh delivery selesai.
  - payload push minimal: `title`, `body`, `issue_id`, `url`, `type`.
  - `url` dibentuk dari `WEB_PUSH_SITE_URL + /issues/{issue_id}` agar klik notifikasi selalu menuju issue publik yang benar.
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
  - `FOLLOWER_TOKEN_SECRET` wajib ada dan minimal 32 karakter; backend memakai secret ini untuk menandatangani `follower_token`.

### Public Stats Dashboard (`GET /api/v1/stats`)

- Endpoint menerima query opsional:
  - `province_id`
  - `regency_id`
- `StatsService` memakai cache in-memory thread-safe (`sync.RWMutex`) dengan TTL 45 detik.
- Cache stats sekarang dibedakan per kombinasi `province_id + regency_id` agar scope wilayah tidak saling tertukar.
- Jika query DB gagal sesaat, service fallback ke cache stale agar endpoint tetap responsif.
- `StatsRepository` menjalankan query agregasi ringan berbasis tabel `issues` + join hirarki `regions`:
  - global stats:
    - `total_issues`
    - `total_issues_this_week`
    - `total_casualties`
    - `total_photos`
    - `total_reports`
  - status stats:
    - `open` dihitung sebagai issue unresolved (`open|verified|in_progress`)
    - `fixed`
    - `archived`
  - time stats:
    - `average_issue_age_days`
    - `oldest_open_issue_age_days`
    - metadata issue tertua unresolved (`oldest_open_issue_id`, lokasi, first seen) jika ada
  - filter metadata:
    - `filters.province_options`
    - `filters.regency_options` untuk provinsi aktif
    - `filters.active_province_id`, `filters.active_regency_id`
    - `filters.scope_label`
    - jika query filter kosong, backend otomatis fallback ke provinsi + kabupaten/kota teratas agar frontend selalu punya default yang masuk akal
  - region leaderboard:
    - daftar kecamatan/subdistrict administratif di scope wilayah aktif
    - ranking berdasarkan `issue_count`, lalu `report_count`, lalu `casualty_count`
    - fallback label terakhir memakai nama region administratif mentah (`raw_region_name`) dan hanya jatuh ke copy generik bila data admin memang kosong
  - top issues:
    - issue dengan laporan terbanyak di scope wilayah aktif
    - issue dengan korban terbanyak di scope wilayah aktif
    - issue paling lama belum diperbaiki di scope wilayah aktif
    - payload issue kini membawa `district_name`, `regency_name`, `province_name` selain `region_name` fallback
- Semua query stats hanya memakai data publik issue:
  - `is_hidden = false`
  - status `rejected` dan `merged` dikecualikan
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
  - jika internal region ketemu: kirim `label`, `region_name`, `region_level`, parent chain, `source=internal_regions`.
  - jika internal region miss tapi reverse geocode berhasil: kirim label fallback manusiawi (`source=reverse_geocode`).
  - jika semua sumber miss: field label bernilai `null`, `source=unresolved`, tanpa memblok submit report.
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
  - log ke `moderation_actions`
  - adjust trust score submitter saat fix/reject

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
  - presign 20/min
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
