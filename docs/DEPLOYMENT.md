# Deployment Guide

## Sumber Deploy Saat Ini

- CI/CD: `.github/workflows/deploy.yml`
- Trigger: push ke branch `main`
- Mekanisme: SSH ke VPS, pull latest, build backend/frontend, restart PM2

## Flow Deploy Aktual

Di workflow saat ini:

1. SSH ke VPS via `appleboy/ssh-action`
2. set `PATH` Node + PM2 binary spesifik user
3. `cd /home/ryandotcodotid/projects/jedug_v2`
4. `git fetch` lalu `git reset --hard origin/main`
5. build backend:
   - `go mod tidy`
   - `go build -o bin/jedug-api ./cmd/api`
   - `pm2 restart jedug-api --update-env`
6. build frontend:
   - `npm ci`
   - `npm run build`
   - `pm2 restart jedug-web --update-env`
7. `pm2 save`

## PM2

Asumsi proses PM2 di server:

- `jedug-api` untuk backend
- `jedug-web` untuk frontend

Catatan: file ecosystem PM2 tidak ada di repo ini.

## Nginx

Konfigurasi Nginx tidak disimpan di repo.

Konsekuensi:

- reverse proxy rules harus divalidasi manual di server
- perubahan domain/SSL/routing tidak ter-track lewat git repo aplikasi

## Env Handling

### Backend env kritikal

- `DATABASE_URL` (required)
- `ADMIN_PASSWORD` (required)
- `APP_PORT`, `CORS_ALLOW_ORIGINS`
- `DUPLICATE_RADIUS_M` (optional, default `30`, satuan meter)
- `STORAGE_DRIVER`, `STORAGE_PUBLIC_BASE_URL`, `UPLOAD_DIR`
- R2 vars saat mode R2 aktif:
  - `R2_ACCESS_KEY_ID`
  - `R2_SECRET_ACCESS_KEY`
  - `R2_BUCKET`
  - `R2_ENDPOINT`
  - `R2_PUBLIC_BASE_URL`

### Frontend env

- `PUBLIC_API_BASE_URL`

## Restart / Rollout Checklist

- pastikan DB reachable dari VPS
- pastikan env backend/frontend up-to-date sebelum restart PM2
- verifikasi endpoint health setelah deploy:
  - `GET /api/v1/health`
- verifikasi UI publik + admin login setelah build frontend

## Current Implementation

- Deploy single workflow, simple, langsung ke VPS.
- Rollout backend/frontend dilakukan dalam satu job.

## Intended Direction

- simpan PM2 ecosystem + template Nginx di repo
- tambahkan smoke-test post-deploy (health + sample API call)
- minimalkan `git reset --hard` untuk mengurangi risiko overwrite file runtime tak terduga

## Known Mismatch

- dokumentasi deployment infra masih tersebar antara workflow dan konfigurasi server manual.
- `.env.example` backend belum mencantumkan semua required variable aktual (`ADMIN_PASSWORD`, vars R2).

## Read This Next

- `docs/ARCHITECTURE.md`
- `docs/STORAGE_AND_MEDIA.md`
- `docs/BACKEND.md`
