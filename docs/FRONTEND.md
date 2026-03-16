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

- `AppHeader.svelte`
  - notification center
  - CTA ringan browser push di dalam panel notifikasi
  - panel `Preferensi Notifikasi` ringan di dropdown yang sama
- `BrowserPushCard.svelte`
  - surface reusable untuk state `unsupported/default/granted/denied/subscribed`
  - dipakai di panel notifikasi dan follow card detail issue
- `NotificationPreferencesPanel.svelte`
  - panel setting ringan untuk master/channel/event toggles
  - tetap memakai notification center, bukan halaman settings terpisah
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
- First load map sekarang memicu geolocate sekali otomatis via kontrol geolocate MapLibre agar blue-dot lokasi user langsung muncul ketika izin lokasi tersedia.
- Geolocate tetap tidak memaksa recenter berulang; tombol geolocate manual tetap dipakai untuk retry.
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
  - timeline vertikal `Riwayat Laporan` (event terbaru di atas)
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
  - loading/error/empty/pagination untuk timeline issue (`GET /api/v1/issues/:id/timeline`)
  - loading/error/following state untuk follow issue anonim (`POST/DELETE /api/v1/issues/:id/follow`)
  - loading/error/status browser push (`unsupported/default/granted/denied/subscribed`)

## Follow Issue Anonim (`/issues/[id]`)

- Halaman detail issue menampilkan card `Ikuti Perkembangan` tanpa menambah login flow baru.
- Identity follower dikelola client-side via localStorage key:
  - `jedug_issue_follower_id`
- Strategy identity:
  - bila belum ada follower id, browser generate UUID via `crypto.randomUUID()` (fallback generator ringan)
  - UUID disimpan stabil per browser/device anonim
- Route detail melakukan:
  - fetch follow status: `GET /api/v1/issues/:id/follow-status?follower_id=...`
  - fetch follower count: `GET /api/v1/issues/:id/followers/count`
  - follow: `POST /api/v1/issues/:id/follow`
  - unfollow: `DELETE /api/v1/issues/:id/follow`
- Semua request follow/status dikirim bersama `X-Device-Token` anonim yang sudah ada dari bootstrap device.
- Response follow/status kini juga bisa membawa:
  - `follower_token`
  - `follower_token_expires_at`
- Frontend menyimpan `follower_token` di localStorage dan memakainya khusus untuk endpoint notification/push.
- Untuk bugfix kompatibilitas, helper frontend follow sekarang memiliki fallback 404 terarah ke alias backend lama/alternatif:
  - status fallback: `/api/v1/issues/:id/follow/status`
  - count fallback: `/api/v1/issues/:id/count`
  - follow/unfollow fallback: `/api/v1/issues/:id/followers`
- UX requirements yang sudah diimplementasikan:
  - disable button saat request berlangsung
  - error copy manusiawi tanpa reload penuh
  - follower count langsung ter-update setelah follow/unfollow berhasil
  - state tetap additive; SSR detail issue tetap berjalan seperti sebelumnya
  - initial load follow tidak lagi menembak dua endpoint sekaligus tanpa alasan; status diprioritaskan dulu, count dipakai sebagai fallback ringan
  - refresh dari klik notifikasi pada issue yang sama juga me-refresh follow state agar count/following status tetap sinkron
  - saat browser ini sudah follow issue, route juga menampilkan CTA `Aktifkan Notifikasi Browser`

## In-App Notification Center

- Notifikasi in-app ditampilkan di `AppHeader` (ikon lonceng + badge unread).
- Data notifikasi dikelola store bersama:
  - `src/lib/stores/notifications.ts`
  - state: `items`, `loading`, `error`, `followerID`, `followerToken`, `initialized`
  - derived: `unreadNotificationCount`
- Preferensi notifikasi dikelola store terpisah:
  - `src/lib/stores/notification-preferences.ts`
  - state utama:
    - `preferences`
    - `loading`
    - `savingKeys`
    - `error`
    - `unavailableMessage`
- Saat app publik pertama kali mount (`routes/+layout.svelte`), frontend menjalankan `notificationsState.init()` setelah bootstrap device.
- Setelah store notif + browser push init, frontend menjalankan `notificationPreferencesState.init()` sekali agar panel settings tidak fetch berulang.
- Jika `follower_token` belum ada/expired, frontend mencoba refresh lewat `POST /api/v1/followers/auth` memakai `X-Device-Token`.
- Endpoint yang dipakai:
  - auth refresh: `POST /api/v1/followers/auth`
  - list: `GET /api/v1/notifications?follower_token=...&limit=50`
  - mark read: `PATCH /api/v1/notifications/:id/read?follower_token=...`
  - delete: `DELETE /api/v1/notifications/:id?follower_token=...`
  - stream realtime: `GET /api/v1/notifications/stream?follower_token=...` (SSE)
  - preferences get: `GET /api/v1/notification-preferences?follower_token=...`
  - preferences patch: `PATCH /api/v1/notification-preferences`
- UX behavior:
  - badge menampilkan jumlah item dengan `read_at = null`
  - dropdown panel menampilkan title/message/waktu + action hapus ringan per item
  - dropdown panel juga menampilkan section `Preferensi Notifikasi` dengan:
    - master switch
    - toggle channel in-app
    - toggle channel push
    - toggle per event type
  - klik item menandai notifikasi sebagai read lalu:
    - navigasi normal ke `/issues/{issue_id}` bila target issue berbeda
    - memicu refresh lokal detail issue + timeline + follow state bila user sudah berada di `/issues/{issue_id}` yang sama
  - delete item menghapus row lokal tanpa full reload; unread badge ikut turun jika item yang dihapus masih unread
  - jika mark-read gagal sinkron ke backend, store menampilkan error ringan dan refresh ulang list dari server agar unread badge tetap konsisten dengan DB
  - jika delete gagal sinkron ke backend, store menampilkan copy `Belum bisa menghapus notifikasi. Coba lagi.` lalu refresh snapshot server agar state tetap akurat
  - read/delete juga disiarkan antar tab via `BroadcastChannel` + fallback `storage` event agar badge/list tetap sinkron
  - reconnect SSE tidak lagi hard-stop setelah beberapa kegagalan; store membersihkan pending timer sebelum reconnect dan melakukan fallback refresh ringan bila gangguan berulang
  - preference update juga disiarkan antar tab via `BroadcastChannel` + fallback `storage` event agar toggle dan policy realtime tetap sinkron
  - saat `notifications_enabled=false` atau `in_app_enabled=false`, store notifikasi memutus SSE realtime agar browser tidak mempertahankan stream yang tidak lagi dipakai
## Browser Push Notification

- Browser push dikelola store terpisah:
  - `src/lib/stores/browser-push.ts`
  - state utama:
    - `status`: `unsupported | default | granted | denied | subscribed`
    - `enabled`
    - `subscribed`
    - `subscriptionCount`
    - `busy`
    - `error` / `success`
- Init flow:
  - dijalankan di `routes/+layout.svelte` setelah `notificationsState.init()`
  - tidak memunculkan permission prompt otomatis
  - hanya memeriksa capability browser, permission saat ini, status backend, dan existing subscription bila permission sudah `granted`
- Endpoint yang dipakai:
  - `POST /api/v1/followers/auth`
  - `GET /api/v1/push/status?follower_token=...`
  - `POST /api/v1/push/subscribe`
  - `POST /api/v1/push/unsubscribe`
- CTA UI:
  - `AppHeader` notification panel memakai `BrowserPushCard.svelte` versi compact
  - detail issue menampilkan `BrowserPushCard.svelte` di bawah follow card setelah browser anonim mengikuti issue
  - panel preferensi menampilkan toggle push, tetapi saat browser push belum aktif toggle dapat dinonaktifkan dan user diarahkan ke `BrowserPushCard`
- Flow enable:
  1. user klik `Aktifkan notifikasi browser`
  2. browser memanggil `Notification.requestPermission()`
  3. bila `granted`, frontend register `/sw.js`
  4. frontend subscribe ke `PushManager` memakai `vapid_public_key` dari backend
  5. subscription dikirim ke backend bersama `follower_token`
- Flow disable:
  1. ambil subscription aktif dari service worker registration
  2. kirim `endpoint` + `follower_token` ke backend
  3. unsubscribe lokal dari browser
- Jika browser belum punya `follower_token` yang valid, CTA enable gagal dengan copy ringan yang mengarahkan user untuk follow issue dulu dari browser yang sama.
- `BrowserPushCard.svelte` sekarang me-refresh `notificationPreferencesState` setelah enable/disable agar toggle push di panel settings tetap sinkron.
- Service worker:
  - file statis: `frontend/static/sw.js`
  - menerima event `push`
  - bila ada tab JEDUG visible:
    - tidak menampilkan OS notification
    - mengirim `postMessage` ke client agar issue aktif bisa di-refresh lokal
  - bila tidak ada tab visible:
    - menampilkan browser notification dengan icon `push-icon.svg`
  - `notificationclick` mencoba fokus issue/tab yang ada dulu, lalu fallback membuka `/issues/{id}`
- Message bridge client:
  - `jedug:push-received` -> refresh issue aktif jika path saat ini sama
  - `jedug:push-open-issue` -> navigate ke issue target atau refresh issue yang sama

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

- payload juga mengirim `actor_follower_id` (identity follower anonim dari browser) untuk mencegah self-notify pada event yang dipicu user itu sendiri.
- backend melakukan normalisasi lokasi saat submit (region internal + reverse geocode road fallback).
- UI tetap menampilkan ini sebagai label UX, bukan input wajib user.

Hardening UX submit report:

- sebelum submit, route `/lapor` selalu memastikan bootstrap device anonim selesai (`ensureDeviceBootstrap`)
- bootstrap memakai guard promise bersama + retry ringan untuk mencegah race condition antar inisialisasi layout dan submit
- bila backend mengembalikan indikasi token bootstrap tidak valid (`device not found; bootstrap first`), frontend melakukan refresh bootstrap sekali lalu retry submit otomatis
- pesan error ke user dipoles agar manusiawi (tidak menampilkan error backend mentah)

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

## Active State Header

- `AppHeader` sekarang memakai route/pathname sebagai source of truth active state.
- Tab `Lapor`, `Peta`, `Statistik` langsung aktif benar pada initial render, refresh, dan navigasi client-side.

## Integrasi API Client

- `lib/api/client.ts`: base fetch helper + ApiError.
- `lib/api/types.ts`: type contracts response backend.
- `lib/api/location.ts`: helper resolve label lokasi `/lapor`.
- `lib/api/stats.ts`: helper fetch dashboard statistik publik `/stats`.
- `lib/api/issues.ts`: helper fetch timeline issue publik (`getIssueTimeline`).
- `lib/api/push.ts`: helper browser push status/subscribe/unsubscribe.
- Token storage:
  - anon token: `jedug_anon_token`
  - issue follower id: `jedug_issue_follower_id`
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
