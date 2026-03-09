# JEDUG Backend

Go + Fiber backend for the JEDUG civic-tech platform.

## Setup

```bash
# 1. Init module (sudah ada di repo, skip kalau sudah ada go.mod)
go mod init jedug_backend

# 2. Install dependencies
go get github.com/gofiber/fiber/v2
go get github.com/jackc/pgx/v5
go get github.com/joho/godotenv
go get github.com/google/uuid
go mod tidy

# 3. Copy env dan isi password DB
cp .env.example .env
# Edit .env → ganti YOUR_PASSWORD dengan password jedug_user

# 4. Jalankan
go run ./cmd/api
```

## Menjalankan

```bash
make run         # dev
make build       # build binary ke bin/jedug-api
make tidy        # go mod tidy
```

## Endpoints

| Method | Path                     | Keterangan                 |
| ------ | ------------------------ | -------------------------- |
| GET    | /api/v1/health           | Health check + DB ping     |
| POST   | /api/v1/device/bootstrap | Bootstrap anonymous device |
| POST   | /api/v1/device/consent   | Simpan consent device      |
| GET    | /api/v1/issues           | List issues publik         |
| GET    | /api/v1/issues/:id       | Detail satu issue          |

## Contoh curl

```bash
# Health check
curl http://localhost:8080/api/v1/health

# Bootstrap device baru
curl -X POST http://localhost:8080/api/v1/device/bootstrap

# Bootstrap device yang sudah ada
curl -X POST http://localhost:8080/api/v1/device/bootstrap \
  -H "X-Device-Token: <anon_token_dari_response_sebelumnya>"

# Consent
curl -X POST http://localhost:8080/api/v1/device/consent \
  -H "Content-Type: application/json" \
  -d '{"anon_token":"<token>","terms_version":"v1.0","privacy_version":"v1.0"}'

# List issues
curl "http://localhost:8080/api/v1/issues?limit=10&offset=0"

# Detail issue
curl http://localhost:8080/api/v1/issues/<uuid>
```

## Asumsi Schema DB

Karena schema tidak dibagikan secara eksplisit, berikut asumsi yang dipakai:

### `devices`

| Kolom        | Tipe                 | Catatan         |
| ------------ | -------------------- | --------------- |
| id           | UUID PK              |                 |
| anon_token   | TEXT UNIQUE NOT NULL |                 |
| user_id      | UUID NULL            | FK ke users     |
| user_agent   | TEXT NULL            |                 |
| ip_address   | INET atau TEXT NULL  | di-cast ke TEXT |
| last_seen_at | TIMESTAMPTZ NOT NULL |                 |
| created_at   | TIMESTAMPTZ NOT NULL |                 |

### `device_consents`

| Kolom           | Tipe                 | Catatan         |
| --------------- | -------------------- | --------------- |
| id              | UUID PK              |                 |
| device_id       | UUID FK              |                 |
| terms_version   | TEXT NOT NULL        |                 |
| privacy_version | TEXT NOT NULL        |                 |
| consented_at    | TIMESTAMPTZ NOT NULL | default = NOW() |

> Kalau kolom di schema-mu adalah `created_at` bukan `consented_at`, ganti di `device_repository.go`.

### `issues`

| Kolom       | Tipe                 | Catatan                            |
| ----------- | -------------------- | ---------------------------------- |
| id          | UUID PK              |                                    |
| title       | TEXT NOT NULL        |                                    |
| description | TEXT NULL            |                                    |
| status      | TEXT NOT NULL        | misal: open, in_progress, resolved |
| is_hidden   | BOOLEAN NOT NULL     | default FALSE                      |
| region_id   | UUID NULL            | FK ke regions                      |
| created_at  | TIMESTAMPTZ NOT NULL |                                    |
| updated_at  | TIMESTAMPTZ NOT NULL |                                    |
