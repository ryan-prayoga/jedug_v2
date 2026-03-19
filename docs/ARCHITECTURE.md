# JEDUG Architecture

Status dokumen ini merefleksikan implementasi aktual per 2026-03-10.

## High-Level Components

- Frontend publik/admin: SvelteKit (`frontend/`)
- Backend API: Go + Fiber (`backend/`)
- Database: PostgreSQL + PostGIS (schema v2)
- Media storage:
  - local filesystem (`UPLOAD_DIR`)
  - Cloudflare R2 (`STORAGE_DRIVER=r2`)
- Deploy: GitHub Actions -> SSH ke VPS -> build + restart PM2

## System Context

1. User membuka frontend dan bootstrap device anonim.
2. Frontend memanggil backend API untuk issue list/detail, upload, dan submit report.
3. Backend menulis data ke Postgres/PostGIS dan media ke storage driver aktif.
4. Admin mengakses route admin untuk moderasi issue/device.

## Arsitektur Backend (Layer)

- `handlers`: parsing request + response mapping
- `service`: business rules (trust, cooldown, moderation policies)
- `repository`: SQL/query + transaksi database
- `storage`: abstraction local/R2 + public URL resolution

Dependency flow: `handler -> service -> repository/storage`.

## Arsitektur Frontend

- `src/routes/*`: page-level flow publik dan admin
- `src/lib/api/*`: client untuk kontrak endpoint
- `src/lib/components/*`: komponen map, sheet, form, state UI
- `src/lib/utils/*`: geolocation, bbox debouncing, image compression, local storage token

## Data + Geo

- Lokasi issue/submission disimpan sebagai `GEOGRAPHY(POINT,4326)`.
- Query map menggunakan bbox + PostGIS spatial operators.
- Smart merge report ke issue existing memakai `ST_DWithin` radius meter (default 30m, configurable), hanya untuk issue aktif publik.

## Flow Kritis Antar Sistem

### A) Submit Report

1. Frontend compress image -> presign upload.
2. Upload file ke R2/local endpoint sesuai `upload_mode`.
3. Frontend submit report dengan `object_key`.
4. Backend validasi device/trust/cooldown.
5. Backend pilih issue aktif terdekat (distance-first + tie-break recency/severity) atau create issue baru.
6. Backend insert issue submission + submission media.

### B) Public Issue Listing (Map)

1. Map emit viewport bbox.
2. Frontend debounce fetch `GET /api/v1/issues?bbox=...`.
3. Backend filter `is_hidden=false` dan status non `rejected/merged`.
4. Frontend render marker + bottom sheet berdasarkan response shape.

### C) Moderation

1. Admin login via env credential.
2. Backend membuat session token in-memory.
3. Admin action update issue/device.
4. Backend catat action ke `moderation_actions`.

## Infra & Operasional

- Workflow deploy tunggal di `.github/workflows/deploy.yml`.
- Deploy backend/frontend di VPS via `gas build` non-interactive.
- Runtime PM2 aktif memakai process `jedug-backend` dan `jedug-frontend`.

## Current Implementation

- Satu backend service monolith + satu frontend app.
- Session admin tidak persisted (memory store).
- Storage migration strategy sudah mempertimbangkan media local lama.
- Schema baseline dan migration additive sekarang versioned di repo (`backend/schema/`, `backend/migrations/`).

## Intended Direction

- Konsolidasi auth admin ke model users/sessions di database.
- Melengkapi logging lifecycle issue (`issue_status_history`) di service.
- Menyatukan config deployment (PM2/Nginx) ke repo agar reproducible.
- Menambahkan migration baru secara additive tanpa mengubah baseline historis secara diam-diam.

## Known Mismatch

- Schema memiliki tabel akun user (users/oauth/sessions), tetapi alur aktif admin masih env + in-memory session.
- Nginx/PM2 config runtime tidak disimpan di repo.
- Sebagian query backend masih defensif terhadap status historis `verified` / `in_progress`, sementara baseline schema issue yang terversion hanya mendefinisikan `open/fixed/archived/rejected/merged`.

## Read This Next

- `docs/BACKEND.md`
- `docs/FRONTEND.md`
- `docs/DEPLOYMENT.md`
- `docs/STORAGE_AND_MEDIA.md`
