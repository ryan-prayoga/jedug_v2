# Database Schema Guide

Dokumen ini disusun dari:

- implementasi query backend (`backend/internal/repository/*.go`)
- baseline schema repo `backend/schema/20260320_000000_baseline.sql`
- migration additive repo `backend/migrations/*.sql`

## Source of Truth di Repo

- Baseline bootstrap penuh:
  - `backend/schema/20260320_000000_baseline.sql`
- Migration additive/idempotent:
  - `backend/migrations/202603140001_create_issue_events.sql`
  - `backend/migrations/202603140002_create_submission_media.sql`
  - `backend/migrations/202603140003_create_issue_followers.sql`
  - `backend/migrations/202603150001_create_notifications.sql`
  - `backend/migrations/202603160001_create_push_subscriptions.sql`
  - `backend/migrations/202603160002_create_follower_auth_bindings.sql`
  - `backend/migrations/202603160003_create_notification_preferences.sql`
  - `backend/migrations/202603160004_create_nearby_alerts.sql`
  - `backend/migrations/202603200001_harden_report_idempotency.sql`
  - `backend/migrations/202603200002_create_push_delivery_jobs.sql`
  - `backend/migrations/202603200003_add_retention_indexes.sql`
  - `backend/migrations/202603200004_create_report_upload_tickets.sql`
  - `backend/migrations/202603200005_persist_submission_admin_location.sql`
- Helper operasional:
  - `backend/scripts/bootstrap_db.sh`
  - `backend/scripts/verify_schema_governance.sh`

## Extensions Wajib

- `postgis`
  - dipakai untuk `GEOGRAPHY(POINT,4326)`, `GEOMETRY(MULTIPOLYGON,4326)`, GiST index, dan fungsi spatial (`ST_DWithin`, `ST_Covers`, `ST_Distance`, `ST_Area`, `ST_X`, `ST_Y`).
- `pgcrypto`
  - dipakai karena code aktif menggunakan `gen_random_uuid()` pada insert `notifications`, `push_subscriptions`, dan nearby alert dispatch path.

## Prinsip Pembacaan

- `issue` = entitas publik kanonik pada peta.
- `issue_submission` = satu laporan mentah dari warga/device terhadap issue.
- `submission_media` = bukti media per submission (bukan langsung per issue).

## Tabel Inti

### `regions`

- Fungsi: master wilayah administratif + geometri polygon.
- Relasi: self-reference `parent_id`; direferensikan oleh `issues` dan `issue_submissions`.
- Kolom penting: `level`, `geom`, `name`, `code`.
- Business meaning: menentukan konteks wilayah issue/submission (terutama district).
- Rawan salah paham: `region_id` bukan mandatory; bisa `NULL` jika lookup region gagal.

### `devices`

- Fungsi: identitas anonim utama pelapor.
- Relasi: parent ke `device_consents`, `issue_submissions`, `issue_flags`, `submission_flags`, `issue_reactions`.
- Kolom penting: `anon_token_hash`, `trust_score`, `is_banned`, `ban_reason`, `last_seen_at`.
- Business meaning: reputasi dan kontrol anti-spam dipusatkan di device.
- Rawan salah paham: token raw tidak disimpan; yang disimpan hash.

### `device_consents`

- Fungsi: rekam persetujuan terms/privacy per device.
- Relasi: FK ke `devices`, opsional ke `users`.
- Kolom penting: `terms_version`, `privacy_version`, `consented_at`, `ip_address`, `user_agent`.
- Business meaning: audit legal consent.
- Rawan salah paham: consent bisa berulang; bukan tabel state tunggal.

### `issues`

- Fungsi: objek publik yang tampil di peta/list/detail.
- Relasi: parent ke `issue_submissions`, `issue_flags`, `issue_reactions`, `issue_daily_stats`, `issue_status_history`.
- Kolom penting:
  - lifecycle: `status`, `verification_status`, `is_hidden`, `hidden_reason`
  - lokasi: `public_location`, `region_id`, `road_name`, `road_type`
  - agregat: `submission_count`, `photo_count`, `casualty_count`, `flag_count`
  - waktu: `first_seen_at`, `last_seen_at`, `resolved_at`
- Business meaning: satu titik masalah jalan yang dikonsolidasikan dari banyak laporan.
- Rawan salah paham:
  - `status` bukan hal yang sama dengan visibility; visibility utama publik tetap dipengaruhi `is_hidden`.
  - `severity_current`/`severity_max` adalah agregasi issue, bukan severity tiap submission.
  - `casualty_count` di issue diperlakukan sebagai nilai casualty tertinggi terlapor pada titik issue (bukan akumulasi mentah semua submission) untuk mencegah overcount laporan duplikat.
  - API publik dapat menurunkan `region_name` via join tabel `regions`; field ini turunan query, bukan kolom fisik di `issues`.

### `issue_submissions`

- Fungsi: laporan mentah per kiriman warga.
- Relasi: FK ke `issues`, `devices`, opsional `users`, `regions`.
- Kolom penting:
  - idempotency: `client_request_id`
  - replay safety: `request_fingerprint`, `created_issue`
  - lokasi: `location`, `region_id`, `district_name`, `regency_name`, `province_name`, `gps_accuracy_m`, `captured_at`, `reported_at`
  - konten: `severity`, `has_casualty`, `casualty_count`, `note`
  - moderasi: `status`, `moderation_note`, `moderated_by`, `moderated_at`
- Business meaning: bukti mentah yang membentuk/menambah issue.
- Constraint/index penting:
  - unique `(device_id, client_request_id)` untuk menjaga idempotency submit per device tanpa bentrok lintas device
- Rawan salah paham:
  - submission bisa `pending` meski issue sudah tampil publik (karena issue adalah agregat).
  - `created_issue=true` menyimpan hasil submit asli agar replay request yang sama bisa mengembalikan `is_new_issue` secara konsisten walau issue tersebut sudah punya submission tambahan setelahnya.
  - `request_fingerprint` dipakai backend untuk menolak reuse `client_request_id` yang payload materialnya berbeda.
  - `district_name` / `regency_name` / `province_name` menyimpan label administratif terbaik yang diketahui saat submit; query issue publik memakainya sebagai fallback ketika `region_id` internal tidak tersedia atau tidak cukup kaya.

### `submission_media`

- Fungsi: file media bukti untuk submission.
- Relasi: FK ke `issue_submissions`.
- Kolom penting: `object_key`, `mime_type`, `size_bytes`, `width`, `height`, `sha256`, `is_primary`, `sort_order`.
- Business meaning: payload media yang dipakai render detail issue.
- Rawan salah paham:
  - backend menganggap media terkait submission, bukan langsung issue.
  - object_key harus key storage, bukan URL mutlak.

### `report_upload_tickets`

- Fungsi: registry pending upload publik sebelum object benar-benar dikonsumsi oleh `submission_media`.
- Relasi: FK ke `devices`; tidak ada FK ke `submission_media` karena row ini sengaja bersifat sementara.
- Kolom penting: `object_key`, `device_id`, `content_type`, `size_bytes`, `upload_mode`, `issued_at`, `expires_at`.
- Constraint/index penting:
  - PK `object_key`
  - index `(device_id, issued_at DESC)` untuk menghitung backlog upload belum dipakai per device
  - index `issued_at` untuk cleanup orphan by age
- Business meaning:
  - satu row = satu ticket upload yang sudah diterbitkan tetapi belum tentu benar-benar dipakai report.
  - service upload memakai tabel ini untuk membatasi churn presign anonim dan memverifikasi bahwa `upload_token` mengacu ke ticket yang memang dikenal backend.
- Rawan salah paham:
  - tabel ini bukan arsip permanen media; submit report yang sukses akan menghapus row pending-nya dalam transaction yang sama.
  - cleanup orphan hanya menyentuh row yang masih tertinggal di tabel ini setelah TTL retention, sehingga media yang sudah terhubung ke report tidak ikut terhapus.

### `issue_flags`

- Fungsi: flag komunitas terhadap issue.
- Relasi: FK ke `issues` dan `devices`; unique `(issue_id, device_id)`.
- Kolom penting: `reason`, `note`, `created_at`.
- Business meaning: sinyal crowd moderation untuk hide/spam control.
- Rawan salah paham: `flag_count` di issue dihitung unique device, bukan jumlah insert mentah.

### `submission_flags`

- Fungsi: flag komunitas terhadap submission spesifik.
- Relasi: FK ke `issue_submissions` dan `devices`.
- Kolom penting: `reason`, `note`.
- Business meaning: granularity moderasi level submission.
- Rawan salah paham: saat ini belum dipakai aktif di flow backend utama.

### `issue_reactions`

- Fungsi: reaksi publik (angry/danger/upvote) per device per issue.
- Relasi: composite PK `(issue_id, device_id)`.
- Kolom penting: `reaction_type`, `created_at`.
- Business meaning: engagement signal issue.
- Rawan salah paham: reaction_count issue belum dikelola service aktif saat ini.

### `issue_status_history`

- Fungsi: jejak perubahan status issue.
- Relasi: FK ke `issues`, opsional ke `users` dan `devices`.
- Kolom penting: `from_status`, `to_status`, `reason`, `created_at`.
- Business meaning: audit trail lifecycle issue.
- Rawan salah paham: tabel tersedia di schema, tetapi saat ini belum diisi oleh service moderasi aktif.

### `issue_events`

- Fungsi: timeline publik riwayat perkembangan issue untuk transparansi.
- Relasi: FK ke `issues` (`ON DELETE CASCADE`).
- Kolom penting: `event_type`, `event_data (JSONB)`, `created_at`.
- Business meaning: jejak event publik lintas lifecycle laporan (pembuatan issue, foto, severity, korban, status).
- Rawan salah paham:
  - `event_data` fleksibel berbasis JSON; shape data bisa berbeda antar `event_type`.
  - timeline publik diurutkan `created_at DESC`, bukan urutan submission id.

### `issue_followers`

- Fungsi: menyimpan subscriber anonim per issue sebagai fondasi notifikasi/update issue.
- Relasi: FK ke `issues` (`ON DELETE CASCADE`).
- Kolom penting: `issue_id`, `follower_id`, `created_at`.
- Constraint/index penting:
  - unique `(issue_id, follower_id)` untuk mencegah follow ganda dari browser/device yang sama
  - index `issue_id` untuk query count follower per issue
  - index `follower_id` untuk lookup daftar issue yang diikuti di langkah berikutnya
- Business meaning: satu browser/device anonim = satu follower ringan tanpa login penuh.
- Rawan salah paham:
  - `follower_id` bukan device token backend; ini identity anonim client-side khusus fitur subscribe.
  - count follower sebaiknya dihitung dari tabel ini langsung sampai nanti benar-benar perlu cache/denormalisasi.

### `notifications`

- Fungsi: daftar notifikasi per-follower yang dibuat otomatis ketika ada event issue baru.
- Relasi: FK ke `issues` (`ON DELETE CASCADE`); `follower_id` adalah UUID anonim client-side (tidak ada FK ke tabel lain).
- Kolom penting: `issue_id`, `follower_id`, `event_id`, `type`, `title`, `message`, `created_at`, `read_at`.
- Constraint/index penting:
  - unique `(event_id, follower_id)` — deduplication agar satu follower tidak menerima notifikasi duplikat untuk event yang sama.
  - index `(follower_id, created_at DESC)` — dipakai backend setelah `follower_token` diverifikasi dan di-resolve ke follower yang sah.
  - index `issue_id` — untuk cleanup cascade.
- Business meaning: follower anonim bisa melihat daftar update issue tanpa login.
- Rawan salah paham:
  - `event_id` merujuk ke `issue_events.id` secara logis, tetapi sengaja belum diberi FK constraint untuk menjaga rollout schema lama tetap aman.
  - dispatch notifikasi berjalan non-fatal setelah event berhasil diinsert; jika gagal, hanya di-log.
  - browser push tidak menggantikan tabel ini; `notifications` tetap source of truth daftar update in-app.
  - `read_at` dipakai aktif oleh endpoint `PATCH /api/v1/notifications/:id/read?follower_token=...` dan menjadi source of truth unread badge di frontend.
- Migration: `backend/migrations/202603150001_create_notifications.sql` — WAJIB DIJALANKAN DI PROD.

### `push_subscriptions`

- Fungsi: menyimpan endpoint browser push aktif per follower anonim.
- Relasi: `follower_id` UUID anonim client-side yang sama dengan sistem follow/notifikasi (tanpa FK ke tabel lain).
- Kolom penting: `follower_id`, `endpoint`, `p256dh`, `auth`, `user_agent`, `created_at`, `updated_at`, `disabled_at`.
- Constraint/index penting:
  - unique `endpoint` — satu subscription browser aktif hanya punya satu row kanonik.
  - index `follower_id`
  - partial index active `follower_id WHERE disabled_at IS NULL`
- Business meaning:
  - satu browser yang memberi izin notifikasi dapat menerima channel tambahan Web Push meski tab JEDUG sedang tidak aktif.
  - subscription bersifat device/browser-scoped, bukan global user account.
- Rawan salah paham:
  - `disabled_at` dipakai untuk soft-disable saat user unsubscribe atau endpoint terbukti invalid/expired (`404/410` dari push service).
  - follower yang sama tetap bisa punya beberapa row historis bila browser merotasi endpoint dari waktu ke waktu.
  - public VAPID key tidak disimpan di tabel; ia berasal dari env backend dan diexpose via endpoint status.
- Migration: `backend/migrations/202603160001_create_push_subscriptions.sql` — WAJIB DIJALANKAN DI PROD.

### `push_delivery_jobs`

- Fungsi: outbox delivery browser push yang tahan restart ringan dan menjadi source of truth status pengiriman async.
- Relasi:
  - FK `issue_id -> issues(id)` (`ON DELETE CASCADE`)
  - `follower_id` tetap UUID anonim client-side yang sama dengan sistem notifikasi (tanpa FK ke tabel lain).
- Kolom penting:
  - payload: `event_id`, `type`, `title`, `message`
  - retry/audit: `attempt_count`, `last_attempt_at`, `next_attempt_at`, `last_error`
  - lifecycle worker: `locked_at`, `delivered_at`, `failed_at`
- Constraint/index penting:
  - unique `(event_id, follower_id)` agar enqueue push tetap idempotent per event notifikasi
  - partial index ready `(next_attempt_at, created_at)` untuk claim worker job yang belum `delivered/failed`
  - check `attempt_count >= 0`
- Business meaning:
  - request utama hanya perlu menulis row outbox; worker background akan mengirim ke provider Web Push sesudah commit.
  - job yang tertinggal saat restart atau crash tetap bisa di-claim ulang karena state-nya persisten di DB.
- Rawan salah paham:
  - tabel ini bukan source of truth daftar notifikasi user; daftar in-app tetap berasal dari `notifications`.
  - `delivered_at` berarti minimal satu subscription aktif follower menerima push dengan sukses; bukan jaminan user melihat notif di OS/browser.
  - `failed_at` berarti retry budget habis atau semua subscription aktif gagal permanen; row sengaja dibiarkan untuk audit.
- Migration: `backend/migrations/202603200002_create_push_delivery_jobs.sql` — WAJIB DIJALANKAN DI PROD.

### `follower_auth_bindings`

- Fungsi: mengikat `follower_id` anonim ke hash `X-Device-Token` sehingga akses notification/push tidak lagi hanya bergantung pada UUID mentah.
- Relasi: `follower_id` UUID anonim client-side yang sama dengan `issue_followers`, `notifications`, dan `push_subscriptions` (tanpa FK ke tabel lain).
- Kolom penting: `follower_id`, `device_token_hash`, `created_at`, `updated_at`.
- Business meaning:
  - backend dapat menerbitkan `follower_token` bertanda tangan server hanya untuk browser yang memegang device token anonim yang benar.
  - notifikasi tetap semi-anonim tanpa menambah sistem login user.
- Rawan salah paham:
  - tabel ini tidak menyimpan raw `X-Device-Token`; hanya hash SHA-256.
  - `follower_token` sendiri tidak disimpan di DB; ia stateless dan diverifikasi via signature + keberadaan binding row.
- Migration: `backend/migrations/202603160002_create_follower_auth_bindings.sql` — WAJIB DIJALANKAN DI PROD.

### `notification_preferences`

- Fungsi: menyimpan preferensi notifikasi minimum per follower anonim tanpa menambah login/account baru.
- Relasi: `follower_id` UUID anonim client-side yang sama dengan `issue_followers`, `notifications`, `push_subscriptions`, dan `follower_auth_bindings` (tanpa FK ke tabel lain).
- Kolom penting:
  - master switch: `notifications_enabled`
  - channel: `in_app_enabled`, `push_enabled`
  - event type: `notify_on_photo_added`, `notify_on_status_updated`, `notify_on_severity_changed`, `notify_on_casualty_reported`, `notify_on_nearby_issue_created`
  - audit: `created_at`, `updated_at`
- Business meaning:
  - follower anonim bisa mengatur channel dan jenis event mana yang masih ingin diterima agar update tidak terasa spammy.
  - nilai child preference tidak dihapus saat `notifications_enabled=false`; master switch hanya membuat semuanya dianggap off sampai user mengaktifkannya lagi.
- Rawan salah paham:
  - `GET /notification-preferences` dapat mengembalikan default sintetis tanpa menulis row; row baru materialized saat user pertama kali menyimpan preference.
  - `push_enabled` boleh `true` meski browser saat ini belum punya subscription aktif; delivery push tetap tidak terjadi tanpa row aktif di `push_subscriptions`.
  - dispatcher notifikasi memakai tabel ini sebagai filter sebelum membuat row `notifications` atau mengantrekan browser push.
- Migration: `backend/migrations/202603160003_create_notification_preferences.sql` — WAJIB DIJALANKAN DI PROD.

### `nearby_alert_subscriptions`

- Fungsi: watched locations anonim per follower/browser untuk memantau issue baru di area tertentu tanpa follow issue satu per satu.
- Relasi: `follower_id` UUID anonim client-side yang sama dengan notification/push/follower auth (tanpa FK ke tabel akun).
- Kolom penting:
  - `label`
  - `latitude`
  - `longitude`
  - `radius_m`
  - `enabled`
  - audit: `created_at`, `updated_at`
- Constraint/index penting:
  - validasi koordinat (`latitude`/`longitude`) di level DB
  - validasi radius `100..5000m`
  - index `(follower_id, updated_at DESC)` untuk panel manajemen
  - GiST expression index geography dari `(longitude, latitude)` untuk lookup radius `ST_DWithin`
- Business meaning:
  - satu follower anonim bisa menyimpan beberapa area seperti rumah/kantor/area kerja.
  - feature ini tetap anonymous-first dan tidak menambah login/account baru.
- Rawan salah paham:
  - `enabled=false` tidak menghapus row; hanya menghentikan matching issue baru.
  - lokasi disimpan eksplisit sebagai lat/lng agar payload CRUD ringan; index spatial tetap dibangun via expression geography.

### `nearby_alert_deliveries`

- Fungsi: dedupe subscription-level untuk memastikan issue baru yang sama tidak mengirim nearby alert berulang ke lokasi pantauan yang sama.
- Relasi:
  - FK `subscription_id -> nearby_alert_subscriptions.id`
  - FK `issue_id -> issues.id`
- Constraint/index penting:
  - unique `(subscription_id, issue_id)` sebagai guard utama anti-duplikasi
  - index `issue_id` untuk audit/debug delivery per issue
  - index `(follower_id, created_at DESC)` untuk histori lightweight bila nanti dibutuhkan
- Business meaning:
  - satu follower bisa punya beberapa lokasi yang overlap; tabel ini memungkinkan backend dedupe per subscription dulu lalu menggabungkan delivery menjadi satu notif per follower.
- Rawan salah paham:
  - row delivery tetap boleh diinsert walau preference/channel sedang off; ini sengaja agar issue lama tidak terkirim retroaktif saat user mengaktifkan setting lagi.
- Migration: `backend/migrations/202603160004_create_nearby_alerts.sql` — WAJIB DIJALANKAN DI PROD.

### `moderation_actions`

- Fungsi: audit log tindakan moderasi admin/system.
- Relasi: opsional FK ke `users` (`actor_user_id`), target polymorphic via `target_type/target_id`.
- Kolom penting: `action_type`, `target_type`, `target_id`, `admin_username`, `note`.
- Business meaning: sumber utama audit tindakan hide/fix/reject/ban/auto-hide.
- Rawan salah paham: actor bisa berasal dari `admin_username` env-based, bukan selalu `actor_user_id`.

### `users`

- Fungsi: model akun user opsional.
- Relasi: parent untuk `oauth_accounts`, `user_sessions`, beberapa FK opsional.
- Kolom penting: `email`, `role`, `is_verified`, `xp_points`, `rank_title`.
- Business meaning: fondasi mode login user/moderator/admin berbasis akun.
- Rawan salah paham: alur aktif saat ini belum memanfaatkan tabel ini untuk admin auth utama.

### `oauth_accounts`

- Fungsi: mapping OAuth provider ke user.
- Relasi: FK ke `users`.
- Kolom penting: `provider`, `provider_user_id`, `provider_email`.
- Business meaning: login sosial (currently `google`).
- Rawan salah paham: belum dipakai di flow runtime aktif.

### `user_sessions`

- Fungsi: sesi refresh token user.
- Relasi: FK ke `users`.
- Kolom penting: `refresh_token_hash`, `expires_at`, `revoked_at`.
- Business meaning: session persistence untuk login akun.
- Rawan salah paham: admin runtime saat ini memakai in-memory session, bukan tabel ini.

### `issue_daily_stats`

- Fungsi: agregat statistik harian issue.
- Relasi: FK ke `issues`, PK `(issue_id, stat_date)`.
- Kolom penting: `views`, `unique_views`, `shares`, `reactions`, `submissions`, `flags`.
- Business meaning: cache statistik/leaderboard/analytics ringan.
- Rawan salah paham: belum dikelola aktif oleh service saat ini.

## Objek Turunan

### `issue_public_view`

- Fungsi: view issue publik dengan turunan `age_days` dan `estimated_loss`.
- Business meaning: penyajian data publik tanpa issue hidden/rejected/merged.
- Rawan salah paham: backend saat ini tidak query view ini langsung; query utama masih ke tabel `issues`.

## Perbedaan Konsep yang Wajib Dipahami

### Issue vs Issue Submission

- `issue`: satu entitas masalah jalan publik pada titik tertentu.
- `issue_submission`: satu kiriman laporan individual untuk issue tersebut.
- Dampak praktis: perubahan severity/status submission tidak otomatis identik dengan status issue kanonik.

### Media Local Lama vs Media R2 Baru

- Local lama: diserve dari `/uploads/gallery/*` pada backend host.
- R2 baru: object_key sama, URL publik via `R2_PUBLIC_BASE_URL`.
- Service storage melakukan fallback ke local jika file legacy masih ada.

### Status Issue vs Moderation Visibility

- Status issue (`open/fixed/archived/rejected/merged`) menjelaskan lifecycle.
- Visibility publik terutama dipengaruhi `is_hidden`.
- Issue bisa status `open` tapi tidak tampil publik jika `is_hidden=true`.

### Device Anonim vs User Login

- Device anonim: identity aktif saat ini untuk submit, flag, trust, ban.
- User login: struktur sudah ada di schema, belum dipakai penuh di flow aktif.

### Severity vs Status

- Severity: tingkat keparahan dampak fisik/bahaya.
- Status: tahap penyelesaian/moderasi issue.
- Issue bisa severity tinggi meski status sudah `fixed` (historical max tetap tinggi).

### `first_seen` vs `last_seen`

- `first_seen_at`: kapan issue pertama kali terdeteksi.
- `last_seen_at`: kapan ada sinyal terbaru (submission/update) terhadap issue.
- Untuk map/list ordering, backend saat ini banyak memakai `last_seen_at DESC`.

## Current Implementation

- Query backend paling banyak memakai tabel:
  - `devices`, `device_consents`, `issues`, `issue_submissions`, `submission_media`, `issue_flags`, `issue_followers`, `moderation_actions`, `issue_events`, `notifications`, `push_subscriptions`, `push_delivery_jobs`
- sebagian tabel schema sudah ada namun belum dipakai penuh (users/oauth/sessions/reactions/submission_flags/daily_stats/history).

## Migration SQL di Repo

- Baseline bootstrap penuh sekarang versioned di repo:
  - `backend/schema/20260320_000000_baseline.sql`
- Migration additive saat ini:
  - `backend/migrations/202603140001_create_issue_events.sql`
  - `backend/migrations/202603140002_create_submission_media.sql`
  - `backend/migrations/202603140003_create_issue_followers.sql`
  - `backend/migrations/202603150001_create_notifications.sql`
  - `backend/migrations/202603160001_create_push_subscriptions.sql`
  - `backend/migrations/202603160002_create_follower_auth_bindings.sql`
  - `backend/migrations/202603160003_create_notification_preferences.sql`
  - `backend/migrations/202603160004_create_nearby_alerts.sql`
  - `backend/migrations/202603200001_harden_report_idempotency.sql`
  - `backend/migrations/202603200002_create_push_delivery_jobs.sql`
  - `backend/migrations/202603200003_add_retention_indexes.sql`
- Index performa yang dipakai timeline:
  - `idx_issue_events_issue_id_created_at (issue_id, created_at DESC, id DESC)`

## Bootstrap dan Verifikasi

- Fresh DB:
  - `cd backend && DATABASE_URL=... ./scripts/bootstrap_db.sh fresh`
- Upgrade DB lama:
  - `cd backend && DATABASE_URL=... ./scripts/bootstrap_db.sh upgrade`
- Verifikasi schema repo vs DB:
  - `cd backend && DATABASE_URL=... ./scripts/verify_schema_governance.sh`

## Known Mismatch dan Verifikasi Manual

- File SQL eksternal historis tidak lagi menjadi source of truth; baseline repo yang baru adalah referensi utama untuk fresh bootstrap.
- File SQL historis memang memiliki typo formatting `submission_media` (`widthINT/heightINT`); baseline dan migration repo sekarang menormalkan kolom menjadi `width` / `height`.
- Sebagian query backend masih memeriksa status issue historis `verified` / `in_progress`; baseline repo tidak memperluas enum issue untuk itu dan menganggap keduanya sebagai mismatch code-level yang perlu dibersihkan terpisah.

## Read This Next

- `docs/BACKEND.md`
- `docs/STORAGE_AND_MEDIA.md`
- `docs/MODERATION.md`
