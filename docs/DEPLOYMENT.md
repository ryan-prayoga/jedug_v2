# Deployment Guide

## Sumber Deploy Saat Ini

- CI/CD: `.github/workflows/deploy.yml`
- Trigger: push ke branch `main`
- Mekanisme: SSH ke VPS, pull latest, deploy backend/frontend via `gas build` non-interactive, lalu verifikasi runtime
- Concurrency guard:
  - GitHub Actions concurrency group `jedug-production-deploy`
  - lock file VPS `${PROJECT_DIR}/.deploy.lock` via `flock`

## Flow Deploy Aktual

Di workflow saat ini:

1. SSH ke VPS via `appleboy/ssh-action`
2. `cd /home/ubuntu/projects/jedug_v2`
3. simpan `PREVIOUS_REF`, lalu `git fetch --prune origin`, `git checkout main`, `git reset --hard origin/main`
4. pastikan PM2 tersedia di sesi non-interactive:

- jika `pm2` belum ada di PATH, workflow akan install user-local via `npm install -g pm2 --prefix ~/.local`

5. jalankan preflight sebelum menyentuh runtime:

- source `backend/.env` dan `frontend/.env` bila file ada
- validasi tools wajib: `gas`, `pm2`, `curl`, `flock`, `git`, `go`, `node`, `psql`, `ss`
- backend:
  - `cd backend && go run ./cmd/preflight`
  - `cd backend && ./scripts/verify_schema_governance.sh`
- frontend:
  - pastikan `PUBLIC_API_BASE_URL` tersedia

6. deploy backend dari `/home/ubuntu/projects/jedug_v2/backend`:

- `gas build --no-ui --yes --type go --pm2-name jedug-backend --port 5000 --git-pull no`

7. deploy frontend dari `/home/ubuntu/projects/jedug_v2/frontend`:

- `gas build --no-ui --yes --type node-web --pm2-name jedug-frontend --port 5001 --git-pull no`

8. verifikasi PM2 status `online` untuk dua proses:

- `jedug-backend`
- `jedug-frontend`

9. verifikasi port dalam kondisi LISTEN:

- `5000` (backend)
- `5001` (frontend)

10. verifikasi healthcheck HTTP nyata:

- backend: `GET http://127.0.0.1:5000/api/v1/health`
- frontend: `GET http://127.0.0.1:5001/health`

11. `pm2 save`
12. jika `gas build` atau runtime verification gagal setelah runtime tersentuh, workflow otomatis rollback ke `PREVIOUS_REF`, rebuild backend+frontend, lalu memverifikasi health lagi

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

## Healthcheck Final Yang Dipakai

### Backend

- Endpoint: `GET /api/v1/health`
- Meaning:
  - process Fiber benar-benar menerima HTTP request
  - koneksi DB aktif (`pgxpool.Ping`)

### Frontend

- Endpoint: `GET /health`
- Meaning:
  - server SvelteKit adapter-node benar-benar serving HTTP
  - `PUBLIC_API_BASE_URL` tersedia
  - frontend bisa menjangkau backend health endpoint yang dikonfigurasi

### Kenapa kombinasi ini dipilih

- `pm2 online` saja tidak cukup karena proses bisa hidup tetapi belum ready.
- `port LISTEN` saja tidak cukup karena socket bisa bind walau app salah config.
- HTTP probe backend menangkap failure runtime + DB reachability.
- HTTP probe frontend menangkap failure runtime + API base mismatch.

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
- upload hardening:
  - `UPLOAD_TOKEN_SECRET` (optional, default fallback ke `FOLLOWER_TOKEN_SECRET`, minimal 32 karakter jika diisi)
  - `UPLOAD_TICKET_TTL_SEC` (optional, default `600`)
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
- auth notifikasi follower:
  - `FOLLOWER_TOKEN_SECRET` (required, minimal 32 karakter random)
  - `FOLLOWER_TOKEN_TTL_SEC` (optional, default `43200`)
  - `FOLLOWER_STREAM_TOKEN_TTL_SEC` (optional, default `600`)

### Database bootstrap / upgrade

- Fresh database baru:
  - `cd backend && DATABASE_URL=... ./scripts/bootstrap_db.sh fresh`
- Existing database yang mengikuti baseline historis:
  - `cd backend && DATABASE_URL=... ./scripts/bootstrap_db.sh upgrade`
- Audit schema setelah apply:
  - `cd backend && DATABASE_URL=... ./scripts/verify_schema_governance.sh`
- Script baseline/migration sudah mengelola extension `postgis` dan `pgcrypto`; pastikan role DB punya privilege `CREATE EXTENSION`.

### Frontend env

- `PUBLIC_API_BASE_URL`

Tidak ada env frontend tambahan untuk VAPID key karena frontend mengambil `vapid_public_key` dari backend lewat `GET /api/v1/push/status`. Frontend juga tidak menyimpan secret follower; ia hanya menyimpan `follower_token` non-SSE dan `stream_token` SSE hasil `POST /api/v1/followers/auth`.

Workflow deploy sekarang memperlakukan `PUBLIC_API_BASE_URL` sebagai env wajib untuk readiness frontend.

## Restart / Rollout Checklist

- pastikan DB reachable dari VPS
- jika deploy ke database baru, jalankan bootstrap schema dari repo sebelum backend start
- jika deploy ke database existing, jalankan upgrade migrations dari repo sebelum backend start
- pastikan env backend/frontend up-to-date sebelum restart PM2
- jalankan preflight backend:
  - `cd backend && go run ./cmd/preflight`
  - `cd backend && ./scripts/verify_schema_governance.sh`
- verifikasi endpoint health setelah deploy:
  - `GET http://127.0.0.1:5000/api/v1/health`
  - `GET http://127.0.0.1:5001/health`
- verifikasi UI publik + admin login setelah build frontend
- verifikasi browser push:
  - buka JEDUG via origin HTTPS
  - follow issue
  - aktifkan notifikasi browser
  - pastikan klik notifikasi membuka `/issues/{id}`
- verifikasi PM2 status backend/frontend = `online`
- verifikasi port `5000` dan `5001` = LISTEN

## Minimum Rollback / Recovery Path

Rollback minimum yang sekarang diotomasi workflow:

1. simpan `PREVIOUS_REF` sebelum pindah ke `origin/main`
2. jika backend/frontend `gas build` gagal atau healthcheck pasca-deploy gagal:
   - `git reset --hard "$PREVIOUS_REF"`
   - rerun `gas build` backend
   - rerun `gas build` frontend
   - ulangi verifikasi PM2 + port + HTTP health
   - `pm2 save`
3. jika rollback juga gagal, workflow fail keras dan butuh intervensi manual di VPS

Recovery manual minimum bila workflow sudah gagal:

1. `cd /home/ubuntu/projects/jedug_v2`
2. `git log --oneline -n 5`
3. `git reset --hard <commit-terakhir-yang-stabil>`
4. `cd backend && gas build --no-ui --yes --type go --pm2-name jedug-backend --port 5000 --git-pull no`
5. `cd ../frontend && gas build --no-ui --yes --type node-web --pm2-name jedug-frontend --port 5001 --git-pull no`
6. cek:
   - `curl -fsS http://127.0.0.1:5000/api/v1/health`
   - `curl -fsS http://127.0.0.1:5001/health`
7. `pm2 save`

## Current Implementation

- Deploy single workflow langsung ke VPS.
- Rollout backend/frontend dilakukan dalam satu job dengan command `gas build` non-interactive.
- Runtime diverifikasi langsung di server lewat PM2 status, port LISTEN, dan HTTP healthcheck backend/frontend.
- Workflow menjalankan preflight env/schema sebelum rollout dan punya rollback minimum ke commit sebelumnya jika rollout gagal di tengah jalan.

## Intended Direction

- simpan PM2 ecosystem + template Nginx di repo
- tambahkan smoke-test post-deploy (health + sample API call)
- minimalkan `git reset --hard` untuk mengurangi risiko overwrite file runtime tak terduga (saat sudah ada deployment strategy yang lebih aman)

## Known Mismatch

- dokumentasi deployment infra masih tersebar antara workflow dan konfigurasi server manual.
- konfigurasi endpoint reverse geocoder production (provider + policy quota) perlu dipastikan sesuai SLA operasional.
- konfigurasi PM2/Nginx runtime masih belum versioned dalam repo.
- rollback masih in-place pada working tree yang sama; ini pragmatis untuk VPS saat ini, tetapi belum sekuat release-directory atomic deploy.

## Read This Next

- `docs/ARCHITECTURE.md`
- `docs/STORAGE_AND_MEDIA.md`
- `docs/BACKEND.md`
