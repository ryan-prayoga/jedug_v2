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
- `GET /api/v1/issues`
- `GET /api/v1/issues/:id`
- `POST /api/v1/issues/:id/flag`

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
- Repository `ReportRepository` (transactional):
  - resolve region district
  - cari issue `open` terdekat dalam 10m
  - create issue baru jika tidak ada
  - create `issue_submissions`
  - create `submission_media`
  - update issue counters

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
  - recent submissions top 3, exclude submission berstatus `rejected`
  - resolve `public_url` media via storage service (compatible local legacy + R2)
  - hanya expose field publik (tanpa device/admin/internal note)

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
