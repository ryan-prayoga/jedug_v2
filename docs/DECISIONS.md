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

## 2026-03-10 - Smart Duplicate Merge ke Issue Aktif (Default 30 Meter)

- Keputusan:
  - submission baru di-attach ke issue aktif publik terdekat (`open|verified|in_progress`, `is_hidden=false`) dalam radius default 30m (configurable via `DUPLICATE_RADIUS_M`).
  - issue `fixed`, `archived`, `rejected`, `merged`, atau hidden tidak dipakai sebagai target merge.
  - jika ada beberapa kandidat, prioritas pemilihan: distance terdekat -> status/verification aktif -> `last_seen_at` terbaru -> severity lebih tinggi.
- Konteks:
  - menurunkan duplikasi marker publik tanpa merusak flow submit yang sudah live.
- Konsekuensi:
  - map lebih bersih dan issue aggregate lebih stabil.
  - laporan dekat issue fixed/archived akan membentuk issue baru (tidak re-open diam-diam).
  - `casualty_count` issue saat merge disimpan sebagai nilai tertinggi terlapor (`GREATEST`) untuk menghindari overcount dari laporan duplikat.

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

## 2026-03-10 - Selective SSR untuk Public Issue Detail

- Keputusan:
  - tetap mempertahankan CSR global di root layout, tetapi mengaktifkan SSR khusus route `/issues/[id]`.
- Konteks:
  - halaman detail issue harus shareable lintas platform sosial dengan metadata yang tersedia saat initial response.
- Konsekuensi:
  - metadata `title/description/og/twitter/canonical` dapat di-generate server-side tanpa merombak flow map publik yang sudah live.
  - implementasi route detail perlu menjaga fallback state client (error/not-found/retry) agar UX tetap stabil saat API bermasalah.

## 2026-03-14 - Anonymous Issue Follow via Browser-Scoped UUID

- Keputusan:
  - fitur follow issue MVP memakai `follower_id` UUID yang di-generate dan disimpan di browser, bukan auth/login baru.
- Konteks:
  - user perlu mengikuti perkembangan issue publik dengan friksi serendah mungkin dan tanpa mengganggu flow anonim yang sudah live.
- Konsekuensi:
  - satu browser/device anonim dianggap satu follower untuk issue tertentu.
  - backend wajib memvalidasi UUID dan memakai unique constraint `(issue_id, follower_id)` agar follow idempotent dan tidak mudah spam.
  - tabel `issue_followers` bisa menjadi pondasi lookup subscriber dan notifikasi di step berikutnya tanpa mengubah kontrak publik yang sudah ada.

## Catatan Governance

- Tambahkan keputusan baru di file ini setiap ada perubahan arsitektur atau kebijakan produk signifikan.
- Referensikan implementasi aktual dan risiko kompatibilitas.

## Read This Next

- `docs/ARCHITECTURE.md`
- `docs/MODERATION.md`
- `docs/STORAGE_AND_MEDIA.md`
