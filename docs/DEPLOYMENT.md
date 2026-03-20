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
- workflow juga memastikan modul `pm2-logrotate` terpasang dan dikonfigurasi:
  - `max_size=10M`
  - `retain=7`
  - `compress=true`
  - rotate harian

5. jalankan preflight sebelum menyentuh runtime:

- source `backend/.env` dan `frontend/.env` bila file ada
- validasi tools wajib: `gas`, `pm2`, `curl`, `flock`, `git`, `go`, `node`, `psql`, `ss`
- backend:
  - `cd backend && go run ./cmd/preflight`
  - `cd backend && ./scripts/verify_schema_governance.sh`
- frontend:
  - pastikan `PUBLIC_APP_BASE_URL` dan `PUBLIC_API_BASE_URL` tersedia

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

10. verifikasi runtime lokal:

- backend: `GET http://127.0.0.1:5000/api/v1/health`
- frontend: `GET http://127.0.0.1:5001/health`

11. verifikasi smoke test publik via ingress/domain:

- frontend: `GET ${PUBLIC_APP_BASE_URL}/health`
- backend health: `GET ${PUBLIC_API_BASE_URL}/api/v1/health`
- sample API publik: `GET ${PUBLIC_API_BASE_URL}/api/v1/issues`
- jalur SSE: `GET ${PUBLIC_API_BASE_URL}/api/v1/notifications/stream` dan **harus** mengembalikan auth error `401/403`, bukan `404/502`

12. `pm2 save`
13. jika `gas build`, verifikasi runtime lokal, atau smoke test publik gagal setelah runtime tersentuh, workflow otomatis rollback ke `PREVIOUS_REF`, rebuild backend+frontend, lalu mengulang verifikasi lokal + publik

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

## Verifikasi Deploy Final Yang Dipakai

### Layer 1 — Runtime Lokal

- backend: `GET http://127.0.0.1:5000/api/v1/health`
- frontend: `GET http://127.0.0.1:5001/health`
- Meaning:
  - proses PM2 tidak hanya `online`, tetapi benar-benar menerima HTTP
  - port backend/frontend tidak hanya bind, tetapi route readiness lokal benar-benar hidup
  - backend local health tetap membuktikan DB + snapshot operasional minimum
  - frontend local health tetap membuktikan server SvelteKit adapter-node siap dan bisa reach backend yang dikonfigurasi

### Layer 2 — Ingress Publik

- frontend health: `GET ${PUBLIC_APP_BASE_URL}/health`
- backend health: `GET ${PUBLIC_API_BASE_URL}/api/v1/health`
- sample API publik: `GET ${PUBLIC_API_BASE_URL}/api/v1/issues`
- SSE path smoke: `GET ${PUBLIC_API_BASE_URL}/api/v1/notifications/stream`
- Meaning:
  - domain/proxy/reverse proxy benar-benar bisa dilalui dari jalur yang sama dengan user publik
  - frontend dan API tidak hanya sehat dari localhost, tetapi reachable lewat origin publik yang dipakai browser
  - route SSE publik minimal terbukti sampai ke backend route yang benar karena mengembalikan auth error terstruktur, bukan proxy/path failure

### Kenapa kombinasi ini dipilih

- `pm2 online` saja tidak cukup karena proses bisa hidup tetapi belum ready.
- `port LISTEN` saja tidak cukup karena socket bisa bind walau app salah config.
- health lokal menangkap failure runtime/process.
- smoke test publik menangkap false-green akibat ingress, reverse proxy, domain, atau path routing yang rusak.
- route SSE publik sengaja tidak mencoba membuat stream auth penuh di deploy workflow agar check tetap stabil; yang diverifikasi adalah reachability path `/api/v1/notifications/stream` lewat proxy.

## Nginx

Template Nginx minimum sekarang disimpan di repo:

- `deploy/nginx/jedug.conf.example`

Konsekuensi:

- runtime file di server tetap perlu di-apply manual sesuai environment
- perubahan domain/SSL/routing penting sekarang punya acuan versioned di repo, tidak lagi sepenuhnya only-on-server

## Nginx — SSE Configuration

Endpoint `GET /api/v1/notifications/stream` memerlukan konfigurasi nginx khusus agar proxy buffering tidak memblokir SSE frames.

Source of truth repo:

- `deploy/nginx/jedug.conf.example`

Jika server memakai file nginx yang berbeda, minimal sinkronkan blok SSE berikut **di dalam** blok `server`, sebelum atau menggantikan blok `/api/` yang lebih umum:

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

Realtime notification sekarang juga membawa `last_event_id` di query string saat reconnect untuk replay ringan. Konfigurasi nginx di atas tetap cukup; tidak ada kebutuhan broker atau endpoint streaming tambahan.

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

- workflow deploy tidak otomatis menulis file nginx aktif di server.
- repo hanya menyimpan template/acuan minimum, bukan memaksa overwrite file runtime server.
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
- `ADMIN_USERNAME` (required, jangan mengandalkan nilai default)
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
  - `UPLOAD_PENDING_WINDOW_SEC` (optional, default `1800`)
  - `UPLOAD_PENDING_LIMIT` (optional, default `4`)
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
- ops / retention:
  - `MAINTENANCE_ENABLED` (optional, default `true`)
  - `MAINTENANCE_INTERVAL_SEC` (optional, default `21600`)
  - `NOTIFICATIONS_RETENTION_DAYS` (optional, default `90`)
  - `PUSH_SUBSCRIPTIONS_STALE_DAYS` (optional, default `180`)
  - `PUSH_SUBSCRIPTIONS_DISABLED_RETENTION_DAYS` (optional, default `30`)
  - `PUSH_DELIVERY_DELIVERED_RETENTION_DAYS` (optional, default `14`)
  - `PUSH_DELIVERY_FAILED_RETENTION_DAYS` (optional, default `30`)
  - `UPLOAD_ORPHAN_RETENTION_SEC` (optional, default `43200`)
- catatan admin cookie auth:
  - jika frontend admin dan backend API berjalan beda origin saat development, `CORS_ALLOW_ORIGINS` harus berupa daftar origin eksplisit; wildcard `*` tidak cukup untuk request `credentials: include`
  - di production via reverse proxy origin yang sama, cookie admin tetap dikirim tanpa storage client-side

### Database bootstrap / upgrade

- Fresh database baru:
  - `cd backend && DATABASE_URL=... ./scripts/bootstrap_db.sh fresh`
- Existing database yang mengikuti baseline historis:
  - `cd backend && DATABASE_URL=... ./scripts/bootstrap_db.sh upgrade`
- Audit schema setelah apply:
  - `cd backend && DATABASE_URL=... ./scripts/verify_schema_governance.sh`
- Script baseline/migration sudah mengelola extension `postgis` dan `pgcrypto`; pastikan role DB punya privilege `CREATE EXTENSION`.

### Frontend env

- `PUBLIC_APP_BASE_URL`
- `PUBLIC_API_BASE_URL`

Tidak ada env frontend tambahan untuk VAPID key karena frontend mengambil `vapid_public_key` dari backend lewat `GET /api/v1/push/status`. Frontend juga tidak menyimpan secret follower; ia hanya menyimpan `follower_token` non-SSE dan `stream_token` SSE hasil `POST /api/v1/followers/auth`.

Workflow deploy sekarang memperlakukan dua env frontend/public sebagai wajib untuk readiness:

- `PUBLIC_APP_BASE_URL`
- `PUBLIC_API_BASE_URL`

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
- verifikasi smoke test publik setelah deploy:
  - `curl -fsS ${PUBLIC_APP_BASE_URL}/health`
  - `curl -fsS ${PUBLIC_API_BASE_URL}/api/v1/health`
  - `curl -fsS ${PUBLIC_API_BASE_URL}/api/v1/issues`
  - `curl -sS -o /tmp/jedug-sse-smoke.out -w '%{http_code}' ${PUBLIC_API_BASE_URL}/api/v1/notifications/stream`
    - expected `401` atau `403`, bukan `404/502`
- verifikasi maintenance manual bila perlu:
  - `cd backend && go run ./cmd/maintenance`
  - cek log `[OPS] maintenance completed ...`
- verifikasi UI publik + admin login setelah build frontend
- verifikasi `POST /api/v1/admin/logout` benar-benar mengeluarkan sesi lama sebelum menganggap deploy berhasil
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
   - ulangi verifikasi PM2 + port + HTTP health lokal
   - ulangi smoke test publik via domain/proxy
   - `pm2 save`
3. jika rollback juga gagal, workflow fail keras dan butuh intervensi manual di VPS
4. jika rollback sukses secara lokal tetapi smoke test publik tetap gagal, anggap masalah ada di ingress/domain/Nginx/runtime config server, bukan sekadar binary app baru

Recovery manual minimum bila workflow sudah gagal:

1. `cd /home/ubuntu/projects/jedug_v2`
2. `git log --oneline -n 5`
3. `git reset --hard <commit-terakhir-yang-stabil>`
4. `cd backend && gas build --no-ui --yes --type go --pm2-name jedug-backend --port 5000 --git-pull no`
5. `cd ../frontend && gas build --no-ui --yes --type node-web --pm2-name jedug-frontend --port 5001 --git-pull no`
6. cek:
   - `curl -fsS http://127.0.0.1:5000/api/v1/health`
   - `curl -fsS http://127.0.0.1:5001/health`
   - `curl -fsS ${PUBLIC_APP_BASE_URL}/health`
   - `curl -fsS ${PUBLIC_API_BASE_URL}/api/v1/health`
   - `curl -fsS ${PUBLIC_API_BASE_URL}/api/v1/issues`
   - `curl -sS -o /tmp/jedug-sse-smoke.out -w '%{http_code}' ${PUBLIC_API_BASE_URL}/api/v1/notifications/stream`
7. `pm2 save`

## Current Implementation

- Deploy single workflow langsung ke VPS.
- Rollout backend/frontend dilakukan dalam satu job dengan command `gas build` non-interactive.
- Deploy sekarang membedakan dua lapis verifikasi:
  - runtime lokal di server
  - smoke test publik via ingress/domain
- PM2 log runtime sekarang diputar otomatis via `pm2-logrotate`; tetap pantau disk VPS, tetapi log tidak lagi tumbuh tanpa batas secara default.
- Workflow menjalankan preflight env/schema sebelum rollout dan punya rollback minimum ke commit sebelumnya jika rollout gagal di tengah jalan.

## Intended Direction

- simpan PM2 ecosystem runtime juga di repo
- minimalkan `git reset --hard` untuk mengurangi risiko overwrite file runtime tak terduga (saat sudah ada deployment strategy yang lebih aman)

## Known Mismatch

- dokumentasi deployment infra masih tersebar antara workflow dan konfigurasi server manual.
- konfigurasi endpoint reverse geocoder production (provider + policy quota) perlu dipastikan sesuai SLA operasional.
- file runtime nginx di server masih manual walau template minimumnya sekarang sudah versioned di repo.
- rollback masih in-place pada working tree yang sama; ini pragmatis untuk VPS saat ini, tetapi belum sekuat release-directory atomic deploy.

## Read This Next

- `docs/ARCHITECTURE.md`
- `docs/STORAGE_AND_MEDIA.md`
- `docs/BACKEND.md`
