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

## Route Admin

- `/admin/login`
- `/admin` (redirect ke `/admin/issues`)
- `/admin/issues`
- `/admin/issues/[id]`

## Komponen Penting

- `IssueMap.svelte`
  - inisialisasi MapLibre
  - marker severity-based
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
- Marker click memilih issue -> tampilkan bottom sheet.
- Reset cache bbox saat user kembali dari list ke map untuk mencegah stuck loading pada viewport yang sama.
- Empty state map hanya dirender setelah fetch viewport valid selesai agar tidak muncul false-empty saat map baru mount.
- Error fetch viewport tidak langsung mengosongkan marker; state terakhir dipertahankan untuk menghindari flicker.
- Empty state map menggunakan info badge top-left (tanpa popup tengah) agar tidak menutupi area peta.

## Detail Issue Publik (`/issues/[id]`)

- Route memakai `+page.ts` (SSR-enabled di level page) untuk:
  - fetch detail issue saat initial request
  - menghasilkan metadata share (`title`, `description`, Open Graph, Twitter card, canonical)
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

## Integrasi Upload + Submit

Di `/lapor`:

1. Ambil lokasi (auto/fallback manual input).
2. Compress image -> WebP.
3. Request presign upload.
4. Upload binary.
   - jika upload R2 gagal, fallback ke endpoint local `/api/v1/uploads/file/{object_key}`
5. Submit report dengan metadata media + `client_request_id`.

## Integrasi API Client

- `lib/api/client.ts`: base fetch helper + ApiError.
- `lib/api/types.ts`: type contracts response backend.
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
