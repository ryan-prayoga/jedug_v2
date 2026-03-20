# JEDUG Backend

Backend API JEDUG dibangun dengan Go + Fiber.

## Quick Start

```bash
cp .env.example .env
make db-fresh
make run
```

Upgrade database lama ke migration chain repo:

```bash
make db-upgrade
make db-verify-schema
```

`make db-fresh`, `make db-bootstrap`, `make db-upgrade`, dan `make db-verify-schema` akan otomatis membaca `backend/.env` bila `DATABASE_URL` belum diexport di shell. `export DATABASE_URL=...` tetap bisa dipakai untuk override sementara.

`make db-fresh` dan `make db-bootstrap` sama-sama menjalankan reset destruktif untuk database target: schema `public` akan di-drop, dibuat ulang, lalu di-bootstrap dari baseline + migration repo.

## Dokumentasi Utama

Agar tidak terjadi duplikasi/inconsistency, gunakan dokumen pusat berikut:

- `AGENTS.md`
- `docs/BACKEND.md`
- `docs/SCHEMA.md`
- `docs/STORAGE_AND_MEDIA.md`
- `docs/MODERATION.md`
- `docs/DEPLOYMENT.md`

## Catatan

`README` ini sengaja dibuat ringkas. Endpoint, schema, dan flow teknis detail dipelihara di folder `docs/` sebagai source of truth lintas agent.
