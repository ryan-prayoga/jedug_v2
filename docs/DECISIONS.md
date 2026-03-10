# Technical and Product Decisions

Format: tanggal - keputusan - konteks - konsekuensi.

## 2026-03-10 - Anonymous-First Identity via Device Token

- Keputusan:
  - pelaporan publik memakai `device` anonim (token hash) sebagai identity utama.
- Konteks:
  - menurunkan friksi onboarding (tanpa login wajib).
- Konsekuensi:
  - trust/ban diarahkan ke device, bukan akun user.
  - raw token hanya muncul di client; backend simpan hash.

## 2026-03-10 - Smart Grouping ke Issue Terdekat (10 Meter)

- Keputusan:
  - submission baru akan di-attach ke issue `open` terdekat dalam radius 10m.
- Konteks:
  - menghindari duplikasi issue pada titik yang sama.
- Konsekuensi:
  - kualitas koordinat dan threshold spasial sangat mempengaruhi akurasi grouping.

## 2026-03-10 - Storage Driver Abstraction (Local dan R2)

- Keputusan:
  - abstraction storage dengan driver aktif (`local`/`r2`) + fallback legacy local media.
- Konteks:
  - migrasi bertahap dari local filesystem ke object storage.
- Konsekuensi:
  - object key convention harus stabil.
  - perubahan URL strategy berisiko merusak media lama jika fallback dihapus.

## 2026-03-10 - Map-First Public Browsing

- Keputusan:
  - halaman publik issue berpusat pada map + bbox query.
- Konteks:
  - discovery issue lebih relevan secara geografis.
- Konsekuensi:
  - response shape `Issue` menjadi kontrak kritis lintas komponen.
  - perubahan field location/status/severity harus sangat hati-hati.

## 2026-03-10 - Admin Auth Sementara via Env Credential + In-Memory Session

- Keputusan:
  - admin auth saat ini memakai env credential dan session memory.
- Konteks:
  - mempercepat delivery moderation MVP.
- Konsekuensi:
  - sesi hilang saat restart service.
  - belum memanfaatkan tabel `users/user_sessions` untuk admin runtime.

## 2026-03-10 - Community Flag Auto-Hide Threshold = 3 Unique Devices

- Keputusan:
  - issue di-hide otomatis ketika unique issue flags >= 3.
- Konteks:
  - respon cepat terhadap konten meragukan tanpa menunggu admin online.
- Konsekuensi:
  - false positive mungkin terjadi di kasus kontroversial.
  - audit auto-hide wajib tercatat di `moderation_actions`.

## Catatan Governance

- Tambahkan keputusan baru di file ini setiap ada perubahan arsitektur atau kebijakan produk signifikan.
- Referensikan implementasi aktual dan risiko kompatibilitas.

## Read This Next

- `docs/ARCHITECTURE.md`
- `docs/MODERATION.md`
- `docs/STORAGE_AND_MEDIA.md`
