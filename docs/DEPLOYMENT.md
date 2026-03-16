# Deployment Guide

## Sumber Deploy Saat Ini

- CI/CD: `.github/workflows/deploy.yml`
- Trigger: push ke branch `main`
- Mekanisme: SSH ke VPS, pull latest, deploy backend/frontend via `gas build` non-interactive, lalu verifikasi runtime

## Flow Deploy Aktual

Di workflow saat ini:

1. SSH ke VPS via `appleboy/ssh-action`
2. `cd /home/ubuntu/projects/jedug_v2`
3. `git fetch --prune origin`, `git checkout main`, `git reset --hard origin/main`
4. pastikan PM2 tersedia di sesi non-interactive:

- jika `pm2` belum ada di PATH, workflow akan install user-local via `npm install -g pm2 --prefix ~/.local`

5. deploy backend dari `/home/ubuntu/projects/jedug_v2/backend`:

- `gas build --no-ui --yes --type go --pm2-name jedug-backend --port 5000 --git-pull no`

6. deploy frontend dari `/home/ubuntu/projects/jedug_v2/frontend`:

- `gas build --no-ui --yes --type node-web --pm2-name jedug-frontend --port 5001 --git-pull no`

7. verifikasi PM2 status `online` untuk dua proses:

- `jedug-backend`
- `jedug-frontend`

8. verifikasi port dalam kondisi LISTEN:

- `5000` (backend)
- `5001` (frontend)

9. `pm2 save`

Semua step wajib fail-fast (`set -Eeuo pipefail`) dengan pesan error jelas agar kegagalan terlihat langsung di log GitHub Actions.

## PM2

Asumsi proses PM2 di server:

- `jedug-backend` untuk backend (port `5000`)
- `jedug-frontend` untuk frontend (port `5001`)

Catatan: file ecosystem PM2 tidak ada di repo ini.

## Command Deploy Non-Interactive (Source of Truth)

Backend:

- `gas build --no-ui --yes --type go --pm2-name jedug-backend --port 5000 --git-pull no`

Frontend:

- `gas build --no-ui --yes --type node-web --pm2-name jedug-frontend --port 5001 --git-pull no`

Alasan tidak menambah flag strategy/install-deps tambahan pada frontend:

- kombinasi `--type node-web` + mode non-interactive (`--no-ui --yes`) sudah cukup untuk flow build/start standar app SvelteKit adapter-node saat ini.
- menghindari coupling ke opsi spesifik yang belum terbukti konsisten lintas versi `gas`.
- menjaga command tetap minimal, eksplisit, dan stabil.

## Nginx

Konfigurasi Nginx tidak disimpan di repo.

Konsekuensi:

- reverse proxy rules harus divalidasi manual di server
- perubahan domain/SSL/routing tidak ter-track lewat git repo aplikasi

## Nginx — SSE Configuration

Endpoint `GET /api/v1/notifications/stream` memerlukan konfigurasi nginx khusus agar proxy buffering tidak memblokir SSE frames.

Tambahkan blok `location` berikut **di dalam** blok `server` nginx, sebelum atau menggantikan blok `/api/` yang lebih umum:

```nginx
# SSE: realtime notification stream
location /api/v1/notifications/stream {
    proxy_pass         http://127.0.0.1:5000;
    proxy_http_version 1.1;

    # Wajib untuk SSE — matikan buffering agar data langsung dikirim ke client
    proxy_buffering    off;
    proxy_cache        off;

    # Timeout panjang agar koneksi SSE tidak diputus nginx
    proxy_read_timeout  3600s;
    proxy_send_timeout  3600s;

    # Header SSE standar
    proxy_set_header Host              $host;
    proxy_set_header X-Real-IP         $remote_addr;
    proxy_set_header X-Forwarded-For   $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_set_header Connection        '';

    # Chunked transfer diperlukan untuk streaming
    chunked_transfer_encoding on;
}
```

> **Catatan**: Header `X-Accel-Buffering: no` juga di-set di response handler Go sebagai
> safeguard tambahan, tapi konfigurasi nginx di atas tetap wajib.

## Browser Push Readiness

Browser push memerlukan syarat runtime tambahan:

- frontend harus disajikan dari origin HTTPS yang sama dengan service worker (`/sw.js`)
- exception development: `http://localhost` tetap valid untuk testing lokal
- `WEB_PUSH_SITE_URL` backend harus menunjuk ke origin frontend publik yang benar, bukan origin API
- file statis `frontend/static/sw.js` dan `frontend/static/push-icon.svg` harus ikut terbawa ke deploy frontend

### Cara Apply Nginx Config

1. Edit file nginx (biasanya `/etc/nginx/sites-available/jedug` atau `/etc/nginx/conf.d/jedug.conf`).
2. Tambahkan blok SSE di atas (sebelum blok `location /api/` generik jika ada).
3. Test: `sudo nginx -t`
4. Reload: `sudo nginx -s reload`

### CI/CD — Apakah Perlu Diubah?

Workflow CI/CD saat ini (`deploy.yml`) sudah cukup untuk deploy code. Nginx config **tidak perlu diubah di CI/CD** karena:

- Nginx config tidak disimpan di repo ini.
- `nginx -s reload` dilakukan sekali manual setelah config baru diapply.
- Perubahan SSE hanya memerlukan restart backend (`gas build` yang sudah ada sudah cukup).

Jika ingin otomasi reload nginx post-deploy, tambahkan step berikut di akhir job deploy:

```yaml
- name: Reload nginx
  run: |
    ssh user@vps "sudo nginx -s reload"
```

## Env Handling

### Backend env kritikal

- `DATABASE_URL` (required)
- `ADMIN_PASSWORD` (required)
- `APP_PORT`, `CORS_ALLOW_ORIGINS`
- `DUPLICATE_RADIUS_M` (optional, default `30`, satuan meter)
- reverse geocode fallback (optional):
  - `REVERSE_GEOCODE_ENABLED` (default `true`)
  - `REVERSE_GEOCODE_URL` (default nominatim reverse endpoint)
  - `REVERSE_GEOCODE_USER_AGENT`
  - `REVERSE_GEOCODE_TIMEOUT_MS` (default `2000`)
  - `REVERSE_GEOCODE_CACHE_TTL_SEC` (default `300`)
- `STORAGE_DRIVER`, `STORAGE_PUBLIC_BASE_URL`, `UPLOAD_DIR`
- R2 vars saat mode R2 aktif:
  - `R2_ACCESS_KEY_ID`
  - `R2_SECRET_ACCESS_KEY`
  - `R2_BUCKET`
  - `R2_ENDPOINT`
  - `R2_PUBLIC_BASE_URL`
- browser push vars saat mode Web Push aktif:
  - `WEB_PUSH_VAPID_PUBLIC_KEY`
  - `WEB_PUSH_VAPID_PRIVATE_KEY`
  - `WEB_PUSH_SUBSCRIBER`
  - `WEB_PUSH_SITE_URL`
  - `WEB_PUSH_TTL_SEC` (optional, default `300`)

### Frontend env

- `PUBLIC_API_BASE_URL`

Tidak ada env frontend tambahan untuk VAPID key karena frontend mengambil `vapid_public_key` dari backend lewat `GET /api/v1/push/status`.

## Restart / Rollout Checklist

- pastikan DB reachable dari VPS
- pastikan env backend/frontend up-to-date sebelum restart PM2
- verifikasi endpoint health setelah deploy:
  - `GET /api/v1/health`
- verifikasi UI publik + admin login setelah build frontend
- verifikasi browser push:
  - buka JEDUG via origin HTTPS
  - follow issue
  - aktifkan notifikasi browser
  - pastikan klik notifikasi membuka `/issues/{id}`
- verifikasi PM2 status backend/frontend = `online`
- verifikasi port `5000` dan `5001` = LISTEN

## Current Implementation

- Deploy single workflow langsung ke VPS.
- Rollout backend/frontend dilakukan dalam satu job dengan command `gas build` non-interactive.
- Runtime diverifikasi langsung di server (PM2 + port check) sebelum job dianggap sukses.

## Intended Direction

- simpan PM2 ecosystem + template Nginx di repo
- tambahkan smoke-test post-deploy (health + sample API call)
- minimalkan `git reset --hard` untuk mengurangi risiko overwrite file runtime tak terduga (saat sudah ada deployment strategy yang lebih aman)

## Known Mismatch

- dokumentasi deployment infra masih tersebar antara workflow dan konfigurasi server manual.
- konfigurasi endpoint reverse geocoder production (provider + policy quota) perlu dipastikan sesuai SLA operasional.
- konfigurasi PM2/Nginx runtime masih belum versioned dalam repo.

## Read This Next

- `docs/ARCHITECTURE.md`
- `docs/STORAGE_AND_MEDIA.md`
- `docs/BACKEND.md`
