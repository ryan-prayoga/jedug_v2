# Frontend Guide

## Stack

- SvelteKit 2 + Svelte 5
- TypeScript
- MapLibre GL untuk peta publik

## Struktur Frontend

- `src/routes/`: halaman publik + admin
- `src/lib/api/`: wrapper fetch dan type contracts
- `src/lib/components/`: UI components
- `src/lib/utils/`: geolocation, bbox, image compression, local storage

## Route Publik

- `/` landing page
- `/lapor` submit report
- `/issues` peta + list issue publik
- `/issues/[id]` detail issue
- `/stats` dashboard statistik agregasi issue publik
- `/api/og/issues/[id]` dynamic Open Graph image generator (PNG 1200x630)

## Route Admin

- `/admin/login`
- `/admin` (redirect ke `/admin/issues`)
- `/admin/issues`
- `/admin/issues/[id]`

## Komponen Penting

- `IssueMap.svelte`
  - inisialisasi MapLibre
  - marker publik via `GeoJSON source + layers`
  - heatmap publik severity-aware via `GeoJSON source + heatmap/circle layers`
  - marker clustering (cluster circle + count + unclustered marker)
  - emit bbox on `moveend`
  - auto geolocation center sekali saat load
- `IssueBottomSheet.svelte`
  - mobile bottom sheet
  - desktop side panel style
  - support swipe/drag down untuk close di mobile (threshold + snap-back)
- `IssueCard.svelte`
  - ringkasan issue untuk list/panel
- `ImagePicker.svelte`
  - image selection + preview
- `ConsentSheet.svelte`
  - consent gate sebelum submit

## Integrasi Map

- `issues/+page.svelte` memanggil `fetchIssuesByBBox`.
- Debounce default 300ms.
- BBox sama (rounded 5 desimal) di-skip untuk mengurangi fetch redundant.
- Route `/issues` sekarang punya mode visual tambahan:
  - `Marker`
  - `Heatmap`
- Marker map publik dirender dari source ter-cluster:
  - cluster circles
  - cluster count labels
  - unclustered hit-area + dot marker severity-based
  - selected marker glow/core layer
- Heatmap dirender dari source terpisah non-cluster agar density tetap stabil:
  - `heatmap density`
  - `heatmap point accent` pada zoom lebih dekat
- Kedua mode memakai payload issue publik yang sama; update data dilakukan via `setData`, bukan re-add source/layer saat toggle.
- Klik cluster melakukan zoom/focus ke area cluster (`getClusterExpansionZoom` + `easeTo`).
- Klik marker individual memilih issue -> tampilkan bottom sheet.
- Saat heatmap aktif, marker individual dan bottom sheet disembunyikan agar peta tetap bersih.
- Klik area peta kosong clear selected issue (dengan guard agar tidak bentrok saat klik cluster/marker).
- Reset cache bbox saat user kembali dari list ke map untuk mencegah stuck loading pada viewport yang sama.
- Empty state map hanya dirender setelah fetch viewport valid selesai agar tidak muncul false-empty saat map baru mount.
- Error fetch viewport tidak langsung mengosongkan marker; state terakhir dipertahankan untuk menghindari flicker.
- Empty state map menggunakan info badge top-left (tanpa popup tengah) agar tidak menutupi area peta.
- Jika setup cluster gagal, komponen fallback ke layer marker individual (tanpa cluster) agar map tetap usable.
- Jika setup heatmap gagal, route otomatis fallback ke mode marker dan menampilkan notice non-blocking; map tidak dipaksa pindah ke list.
- Formula weight heatmap saat ini:
  - base severity: level 1 `0.18`, 2 `0.34`, 3 `0.58`, 4 `0.78`, 5+ `0.92`
  - bonus casualty: `+0.06` per korban sampai 3
  - bonus submission: `+0.02` per laporan tambahan sampai 4 laporan tambahan
  - status multiplier: `0.45` untuk `fixed/archived`, `1` untuk issue aktif
  - final weight di-clamp ke `0.08..1.00`

## Detail Issue Publik (`/issues/[id]`)

- Route memakai `+page.ts` (SSR-enabled di level page) untuk:
  - fetch detail issue saat initial request
  - menghasilkan metadata share (`title`, `description`, Open Graph, Twitter card, canonical)
  - mengarah ke OG image dinamis per issue: `/api/og/issues/{id}`
  - treat `400/404` sebagai not-found publik
- UI detail page bersifat mobile-first:
  - hero media + fallback placeholder
  - meta issue (severity/status/verification/lokasi/first seen/last seen)
  - metrik ringkas (laporan/foto/korban/reaksi/update terakhir)
  - galeri media publik sederhana + preview lightbox
  - detail tambahan + `public_note` ringkas + aktivitas terbaru yang memakai `recent_submissions[].public_note`
  - CTA share + social links + open external map
- Komponen route dipisah agar maintainable:
  - `IssueHeader.svelte`
  - `IssueStats.svelte`
  - `IssueGallery.svelte`
  - `ShareActions.svelte`
- Layout desktop issue detail memakai container lebar sendiri (`app-main-wide`) tanpa mengubah flow route publik lain.
- State wajib tersedia:
  - loading (retry fetch)
  - not found
  - error
  - fallback media gagal load
  - empty gallery

## Public Stats Dashboard (`/stats`)

- Halaman mobile-first untuk civic storytelling berbasis agregasi data issue publik.
- Route melakukan fetch `GET /api/v1/stats` via `lib/api/stats.ts`.
- Struktur section:
  - Global Stats (card grid)
  - Status Breakdown (card + progress bar)
  - Time Stats (rata-rata umur issue + issue tertua unresolved)
  - Region Leaderboard (list berurutan)
  - Top Issue (card list dengan link ke `/issues/[id]`)
- State wajib:
  - loading
  - error + retry
  - empty
- Halaman tetap memakai komponen state umum:
  - `LoadingState`
  - `ErrorState`
  - `EmptyState`
- Data sensitif tidak ditampilkan di UI stats karena endpoint hanya expose agregasi publik.

## Dynamic OG Image (`/api/og/issues/[id]`)

- Endpoint server route SvelteKit (`+server.ts`) yang merender `image/png` ukuran `1200x630`.
- Source data OG mengikuti endpoint backend publik `GET /api/v1/issues/:id`.
- Jika issue punya foto (`primary_media` atau media pertama), foto dipakai sebagai background + dark overlay.
- Jika issue tidak punya foto, endpoint pakai gradient brand (`#E5484D`) sebagai fallback.
- Jika issue tidak ditemukan / API gagal, endpoint mengembalikan fallback OG image PNG (bukan JSON error) agar crawler sosial tetap dapat preview.
- Header cache diset aman untuk crawler/CDN:
  - response issue valid: short-lived public cache + `stale-while-revalidate`
  - fallback response: cache lebih pendek

## Integrasi Upload + Submit

Di `/lapor`:

1. Ambil lokasi (auto/fallback manual input).
2. Resolve label lokasi manusiawi via endpoint internal `GET /api/v1/location/label`.
   - dipanggil saat lokasi berhasil didapat atau koordinat manual dipilih eksplisit
   - ada cache in-memory berbasis koordinat ter-rounded untuk menghindari request berulang
   - gagal resolve label tidak memblok submit (koordinat tetap dipakai)
3. Compress image -> WebP.
4. Request presign upload.
5. Upload binary.
   - jika upload R2 gagal, fallback ke endpoint local `/api/v1/uploads/file/{object_key}`
6. Submit report dengan metadata media + `client_request_id`.
   - backend melakukan normalisasi lokasi saat submit (region internal + reverse geocode road fallback).
   - UI tetap menampilkan ini sebagai label UX, bukan input wajib user.

## UX Lokasi di `/lapor`

- Panel lokasi menampilkan:
  - koordinat mentah (`lat, lon`) sebagai acuan utama
  - label lokasi manusiawi (primary + secondary) dari endpoint `/api/v1/location/label` (internal region + reverse fallback)
  - helper text bahwa nama jalan dilengkapi otomatis saat laporan dikirim
- Bila label wilayah tidak tersedia:
  - user tetap bisa submit report
  - koordinat tetap dipakai penuh oleh backend.

## Entry Point Statistik

- Navigasi header publik sekarang menyediakan link:
  - `Lapor`
  - `Peta`
  - `Statistik`
- Homepage juga menyediakan CTA langsung ke `/stats` agar dashboard mudah ditemukan tanpa mengetik URL manual.

## Integrasi API Client

- `lib/api/client.ts`: base fetch helper + ApiError.
- `lib/api/types.ts`: type contracts response backend.
- `lib/api/location.ts`: helper resolve label lokasi `/lapor`.
- `lib/api/stats.ts`: helper fetch dashboard statistik publik `/stats`.
- Token storage:
  - anon token: `jedug_anon_token`
  - admin token: `jedug_admin_token`

## Area Sensitif Jangan Dirusak

- shape `Issue` object dan naming field snake_case dari backend
- alur bbox fetch + map marker render
- alur upload fallback (R2 -> local endpoint)
- consent bootstrap flow sebelum submit report
- auth guard admin layout (`/admin/+layout.svelte`)
- metadata SSR untuk `/issues/[id]` (title/description/OG/Twitter/canonical)
- additive field detail issue (`primary_media`, `public_note`, `recent_submissions[].public_note`) diprioritaskan untuk halaman publik agar tidak menampilkan note mentah

## Current Implementation

- Public UI sudah map-first dan cukup konsisten dengan design-docs.
- Admin UI fungsional untuk moderation utama.
- SSR global tetap nonaktif di root layout, tetapi detail issue publik `/issues/[id]` diaktifkan SSR pada page-level untuk shareability.

## Known Mismatch

- Design polishing fokus besar di public UI; admin UI belum sepenuhnya mengikuti style token terbaru.
- Sebagian mapping label status di UI tidak eksplisit mencakup semua enum schema (`merged`, `rejected`, dll), sehingga fallback ke raw status string bisa muncul.

## Read This Next

- `docs/MAP_AND_LOCATION.md`
- `docs/STORAGE_AND_MEDIA.md`
- `docs/MODERATION.md`
- `docs/DESIGN_INDEX.md`
