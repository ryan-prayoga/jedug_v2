# Database Schema Guide

Dokumen ini disusun dari:

- implementasi query backend (`backend/internal/repository/*.go`)
- schema SQL v2 yang diberikan di `/Users/ryanprayoga/Downloads/jedug_schema_v2.sql`

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
  - lokasi: `location`, `gps_accuracy_m`, `captured_at`, `reported_at`
  - konten: `severity`, `has_casualty`, `casualty_count`, `note`
  - moderasi: `status`, `moderation_note`, `moderated_by`, `moderated_at`
- Business meaning: bukti mentah yang membentuk/menambah issue.
- Rawan salah paham: submission bisa `pending` meski issue sudah tampil publik (karena issue adalah agregat).

### `submission_media`

- Fungsi: file media bukti untuk submission.
- Relasi: FK ke `issue_submissions`.
- Kolom penting: `object_key`, `mime_type`, `size_bytes`, `width`, `height`, `sha256`, `is_primary`, `sort_order`.
- Business meaning: payload media yang dipakai render detail issue.
- Rawan salah paham:
  - backend menganggap media terkait submission, bukan langsung issue.
  - object_key harus key storage, bukan URL mutlak.

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
  - index `(follower_id, created_at DESC)` — dipakai oleh GET `/api/v1/notifications?follower_id=...`.
  - index `issue_id` — untuk cleanup cascade.
- Business meaning: follower anonim bisa melihat daftar update issue tanpa login.
- Rawan salah paham:
  - `event_id` merujuk ke `issue_events.id` secara logis, tetapi tidak ada FK constraint karena tipe kolom belum diverifikasi eksplisit.
  - dispatch notifikasi berjalan non-fatal setelah event berhasil diinsert; jika gagal, hanya di-log.
  - `read_at` belum dipakai aktif di API (disimpan untuk fitur mark-as-read di masa depan).
- Migration: `backend/migrations/202603150001_create_notifications.sql` — WAJIB DIJALANKAN DI PROD.

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
  - `devices`, `device_consents`, `issues`, `issue_submissions`, `submission_media`, `issue_flags`, `issue_followers`, `moderation_actions`, `issue_events`, `notifications`
- sebagian tabel schema sudah ada namun belum dipakai penuh (users/oauth/sessions/reactions/submission_flags/daily_stats/history).

## Migration SQL di Repo

- Timeline issue events sekarang memiliki migration versioned di repo:
  - `backend/migrations/202603140001_create_issue_events.sql`
- Index performa yang dipakai timeline:
  - `idx_issue_events_issue_id_created_at (issue_id, created_at DESC, id DESC)`

## Known Mismatch dan Verifikasi Manual

- SQL source saat ini file eksternal di luar repo; perlu dipindah ke repo agar versioned.
- Ditemukan indikasi formatting typo pada SQL `submission_media` (`widthINT/heightINT`) yang perlu diverifikasi terhadap schema DB aktual.
- Backend mengandalkan kolom `width`/`height`; pastikan schema database nyata memang memiliki nama kolom tersebut.

## Read This Next

- `docs/BACKEND.md`
- `docs/STORAGE_AND_MEDIA.md`
- `docs/MODERATION.md`
