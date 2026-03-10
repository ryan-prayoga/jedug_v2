# Moderation Guide

## Tujuan Moderasi

- menjaga kualitas issue publik
- menurunkan spam/hoax/duplikasi
- menjaga trust score submitter
- menyediakan jejak audit moderasi

## Admin Login Flow

Implementasi aktif:

- login pakai env credential (`ADMIN_USERNAME`, `ADMIN_PASSWORD`)
- endpoint: `POST /api/v1/admin/login`
- session token disimpan in-memory di backend
- endpoint protected memakai `Authorization: Bearer <token>`

## Moderation Actions

Endpoint utama:

- `POST /api/v1/admin/issues/:id/hide`
- `POST /api/v1/admin/issues/:id/unhide`
- `POST /api/v1/admin/issues/:id/fix`
- `POST /api/v1/admin/issues/:id/reject`
- `POST /api/v1/admin/devices/:id/ban`

Efek bisnis:

- hide/unhide: kontrol visibility publik (`is_hidden`)
- fix/reject: ubah status issue + adjust trust score submitter
- ban device: `is_banned=true`, `trust_score=-100`

## Community Moderation (Flag)

- endpoint: `POST /api/v1/issues/:id/flag`
- reason valid: `spam|hoax|off_topic|duplicate|abuse|other`
- satu device hanya bisa flag satu issue satu kali (unique constraint)
- auto-hide issue saat unique flag count >= 3
- auto-hide action dicatat sebagai `moderation_actions` oleh `system`

## Moderation Log

- tabel audit: `moderation_actions`
- admin issue detail menampilkan log terbaru
- field penting:
  - `action_type`
  - `target_type`
  - `target_id`
  - `admin_username` / actor
  - `note`

## Trust / Hardening Relevan

- trust score digunakan saat submit report:
  - device dengan trust terlalu rendah ditolak
- cooldown submit 2 menit/device
- rate limit per endpoint
- banned device ditolak untuk submit report dan flag

## Current Implementation

- Moderasi issue/device sudah end-to-end dari UI admin ke backend.
- Auto-hide berbasis komunitas sudah aktif.
- Logging action moderasi sudah ada.

## Intended Direction

- satukan actor moderasi ke model user DB (bukan env-only admin)
- tambah action history lifecycle issue di `issue_status_history`
- perluas tooling review submission-level moderation (`submission_flags`)

## Known Mismatch

- schema mendukung actor user (`actor_user_id`), tetapi implementasi aktif lebih sering mengisi `admin_username`.
- session admin tidak persisted ke database (`user_sessions` belum dipakai).

## Read This Next

- `docs/BACKEND.md`
- `docs/SCHEMA.md`
- `docs/FRONTEND.md`
