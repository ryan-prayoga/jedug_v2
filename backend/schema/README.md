## Database Schema Source of Truth

Gunakan artefak berikut sebagai source of truth schema di repo:

- `backend/schema/20260320_000000_baseline.sql`
  Baseline schema penuh untuk bootstrap database baru.
- `backend/migrations/*.sql`
  Migration additive/idempotent untuk upgrade dari baseline historis yang sebelumnya hidup di file eksternal.
- `backend/scripts/bootstrap_db.sh`
  Helper untuk bootstrap fresh DB atau apply upgrade migrations.
- `backend/scripts/verify_schema_governance.sh`
  Helper audit extension, tabel, kolom, dan index penting yang diasumsikan code.

Current implementation:

- baseline SQL sudah dicatat penuh di repo dan tidak lagi bergantung pada file di luar workspace.
- mode `fresh` pada `backend/scripts/bootstrap_db.sh` sekarang mereset schema `public` lebih dulu (`DROP SCHEMA public CASCADE` -> create ulang schema) sebelum baseline + migration dijalankan, sehingga bootstrap benar-benar dimulai dari database kosong.
- migration chain additive tetap tersedia agar environment lama bisa dikejar ke schema yang diharapkan code saat ini.

Intended direction:

- fresh deploy memakai baseline repo.
- perubahan schema berikutnya ditambahkan sebagai migration baru yang idempotent, lalu docs disinkronkan.

Known mismatch:

- sebagian query backend masih defensif terhadap status historis `verified` / `in_progress`, sementara baseline schema issue tetap mendefinisikan status kanonik `open/fixed/archived/rejected/merged`.
- tabel `notifications.event_id` sengaja belum diberi FK ke `issue_events.id` agar rollout ke environment lama tetap aman; relasi saat ini logical-only.
