# Storage and Media Guide

## Gambaran Umum

JEDUG mendukung dua mode storage:

- `local` (filesystem server)
- `r2` (Cloudflare R2, S3-compatible)

Pemilihan mode dikendalikan oleh env `STORAGE_DRIVER`.

## Object Key Convention

Format object key wajib:

- `issues/YYYY/MM/<uuid>.<ext>`

Aturan validasi:

- harus lowercase
- tidak boleh absolute URL
- tidak boleh mengandung `..` atau spasi
- extension harus salah satu: `.jpg/.jpeg/.png/.webp/.heic/.heif`

## Upload Flow (Presign -> Upload -> Submit)

1. Frontend request `POST /api/v1/uploads/presign` dengan `anon_token`.
2. Backend validasi device anonim, lalu generate:
   - `object_key` yang fixed dari server
   - upload target berdasarkan driver aktif
   - `upload_token` bertanda tangan server untuk satu file report (`device_id + object_key + mime_type + size_bytes + expiry`)
3. Frontend upload binary:
   - local mode -> `POST /api/v1/uploads/file/{object_key}` dengan header `X-Upload-Token`
   - r2 mode -> `PUT` ke presigned URL R2
4. Frontend submit report ke `/api/v1/reports` dengan metadata media + `upload_token`.
5. Backend menolak report jika:
   - `upload_token` invalid/expired
   - device submit tidak cocok dengan owner ticket
   - object belum ada di storage
   - ukuran / mime aktual tidak cocok dengan ticket
   - `object_key` sudah pernah dipakai di report lain

## Local vs R2 Behavior

### Local

- Public URL dibangun dari `STORAGE_PUBLIC_BASE_URL + /uploads/gallery/{object_key}`
- File ditulis ke `UPLOAD_DIR`
- Endpoint static:
  - `/uploads/gallery/*` diserve langsung oleh Fiber
- Jalur upload binary tidak lagi publik murni; wajib `X-Upload-Token` yang valid.

### R2

- Presigned URL via AWS SDK S3 client
- Public URL dari `R2_PUBLIC_BASE_URL/{object_key}`
- Upload langsung ke bucket R2
- Validasi ownership tidak berhenti di presign URL; `/reports` tetap memverifikasi `upload_token` + keberadaan object di bucket.

## Public URL Strategy

`storage.Service.ResolvePublicURL` punya strategi penting:

1. jika field media sudah absolute URL -> pakai langsung
2. jika file object key masih ada di local legacy -> gunakan URL local
3. selain itu -> gunakan active driver URL builder

Ini menjaga backward compatibility saat migrasi local -> R2.

## Backward Compatibility Media Lama

Backend tetap mount static local path meskipun mode aktif `r2`.

Tujuannya:

- media lama yang dulu tersimpan lokal tetap bisa diakses
- issue detail lama tidak rusak saat driver berpindah

## Validasi dan Hardening Media

- max file size: 20 MB
- MIME whitelist ketat
- object key + extension harus cocok dengan mime type
- body upload kosong ditolak
- upload ticket expiry pendek (default 10 menit)
- caller tidak bisa menentukan `object_key` sendiri
- local upload tanpa `X-Upload-Token` ditolak
- media hanya bisa dipakai sekali lintas `submission_media`

## Current Implementation

- Storage abstraction sudah rapi dan teruji (`storage_test.go`).
- Frontend lapor tetap punya fallback upload saat presign R2 gagal, tetapi fallback local kini ikut membawa `X-Upload-Token`.

## Known Mismatch

- Konvensi object key saat ini fixed untuk prefix `issues/`; jika nanti multi-entity media dibutuhkan, perlu revisi kontrak dan migrasi.
- SQL schema source belum berada di repo, sehingga governance media fields (`width`, `height`, `metadata`) perlu verifikasi manual lintas environment.

## Read This Next

- `docs/BACKEND.md`
- `docs/FRONTEND.md`
- `docs/SCHEMA.md`
- `docs/DEPLOYMENT.md`
