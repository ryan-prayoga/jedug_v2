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
- `/health` readiness route server-side untuk deploy/runtime verification; jalur publik ini sekarang dipakai workflow smoke test melalui `PUBLIC_APP_BASE_URL`
- `/api/og/issues/[id]` dynamic Open Graph image generator (PNG 1200x630)

## Route Admin

- `/admin/login`
- `/admin` (redirect ke `/admin/issues`)
- `/admin/issues`
- `/admin/issues/[id]`
- admin frontend tidak lagi menyimpan bearer token di `localStorage`
- helper `src/lib/api/admin.ts` sekarang selalu memakai `credentials: 'include'` agar browser mengirim cookie session admin
- logout dilakukan lewat `POST /api/v1/admin/logout`, bukan sekadar menghapus state client
- shell admin kini aman pada initial load/hard refresh route nested (`/admin/issues/[id]`); pengecekan sesi tidak lagi crash saat `afterNavigate.from` bernilai `null`

## Komponen Penting

- `AppHeader.svelte`
  - notification center
  - CTA ringan browser push di dalam panel notifikasi
  - panel `Preferensi Notifikasi` ringan di dropdown yang sama
  - panel `Nearby Alerts` ringan di dropdown yang sama untuk mengelola area pantauan
- `BrowserPushCard.svelte`
  - surface reusable untuk state `unsupported/default/granted/denied/subscribed`
  - dipakai di panel notifikasi dan follow card detail issue
- `NotificationPreferencesPanel.svelte`
  - panel setting ringan untuk master/channel/event toggles
  - tetap memakai notification center, bukan halaman settings terpisah
- `NearbyAlertsPanel.svelte`
  - panel ringan untuk create/list/update/delete watched locations
  - lazy-load data hanya saat panel dibuka agar app init tetap ringan
  - mendukung autofill koordinat dari geolocation browser + input manual
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
  - meta issue (severity/status/verification/wilayah/first seen/last seen)
  - metrik ringkas (laporan/foto/korban/reaksi)
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
- Presenter lokasi issue publik sekarang memakai prioritas yang lebih manusiawi:
  - `road_name` bila benar-benar nama jalan/area
  - fallback ke `district_name/regency_name/province_name` bila `road_name` lama hanya berisi copy koordinat sintetis
  - fallback koordinat hanya dipakai jika semua label manusiawi kosong
- Koordinat issue detail sekarang diperlakukan sebagai informasi sekunder:
  - tidak lagi diulang di hero + ringkasan + aside sekaligus
  - tetap tersedia jelas di kartu lokasi samping
- `Tipe jalan` hanya ditampilkan bila backend memang mengirim nilai yang layak pakai; field kosong tidak lagi dirender sebagai noise `Belum tersedia`.
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
  - `follower_stream_token`
  - `follower_stream_token_expires_at`
- Frontend menyimpan dua token di localStorage:
  - `follower_token` untuk endpoint notification/push non-SSE
  - `follower_stream_token` khusus koneksi SSE
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
- Nearby alerts dikelola store terpisah:
  - `src/lib/stores/nearby-alerts.ts`
  - state utama:
    - `items`
    - `loading`
    - `creating`
    - `savingKeys`
    - `deletingIDs`
    - `error`
    - `unavailableMessage`
- Saat app publik pertama kali mount (`routes/+layout.svelte`), frontend menjalankan `notificationsState.init()` setelah bootstrap device.
- Setelah store notif + browser push init, frontend menjalankan `notificationPreferencesState.init()` sekali agar panel settings tidak fetch berulang.
- Nearby alerts tidak ikut di-init pada first paint; panel melakukan fetch lazy saat user benar-benar membuka section tersebut.
- Jika `follower_token` belum ada/expired, frontend mencoba refresh lewat `POST /api/v1/followers/auth` memakai `X-Device-Token`.
- Endpoint yang dipakai:
  - auth refresh: `POST /api/v1/followers/auth`
  - list: `GET /api/v1/notifications?limit=50` dengan `X-Follower-Token` + `X-Device-Token`
  - mark read: `PATCH /api/v1/notifications/:id/read` dengan `X-Follower-Token` + `X-Device-Token`
  - delete: `DELETE /api/v1/notifications/:id` dengan `X-Follower-Token` + `X-Device-Token`
  - stream realtime: `GET /api/v1/notifications/stream?stream_token=...&last_event_id=...` (SSE)
  - preferences get: `GET /api/v1/notification-preferences` dengan `X-Follower-Token` + `X-Device-Token`
  - preferences patch: `PATCH /api/v1/notification-preferences`
  - nearby alerts list: `GET /api/v1/nearby-alerts` dengan `X-Follower-Token` + `X-Device-Token`
  - nearby alerts create: `POST /api/v1/nearby-alerts`
  - nearby alerts patch: `PATCH /api/v1/nearby-alerts/:id`
  - nearby alerts delete: `DELETE /api/v1/nearby-alerts/:id`
- Frontend tidak lagi menaruh token notification umum di query string untuk endpoint non-SSE; query string kini dipakai hanya untuk `stream_token` SSE yang TTL-nya pendek.
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
  - store menyimpan cursor `event_id` terbesar dari snapshot atau SSE
  - reconnect SSE mengirim `last_event_id` agar backend bisa menutup gap pendek lewat replay ringan
  - reconnect tidak lagi hanya menunggu event berikutnya; setelah reconnect/tab resume, store menjalankan reconciliation snapshot ringan agar badge/list tidak stale
  - reconciliation periodik ringan saat tab visible dipakai untuk menutup gap akibat sleep/background atau buffer drop tanpa membuat sistem replay besar
  - preference update juga disiarkan antar tab via `BroadcastChannel` + fallback `storage` event agar toggle dan policy realtime tetap sinkron
  - saat `notifications_enabled=false` atau `in_app_enabled=false`, store notifikasi memutus SSE realtime agar browser tidak mempertahankan stream yang tidak lagi dipakai

## Browser Push Notification

- Browser push dikelola store terpisah:
  - `src/lib/stores/browser-push.ts`
  - state utama:
    - `status`: `ios_browser_tab | unsupported | default | granted | denied | subscribed`
    - `enabled`
    - `subscribed`
    - `subscriptionCount`
    - `busy`
    - `error` / `success`
- Init flow:
  - dijalankan di `routes/+layout.svelte` setelah `notificationsState.init()`
  - tidak memunculkan permission prompt otomatis
  - memeriksa capability browser, permission saat ini, status backend, existing subscription bila permission sudah `granted`, serta mode iOS `Home Screen` vs tab browser biasa
- Endpoint yang dipakai:
  - `POST /api/v1/followers/auth`
  - `GET /api/v1/push/status` dengan `X-Follower-Token` + `X-Device-Token`
  - `POST /api/v1/push/subscribe`
  - `POST /api/v1/push/unsubscribe`
- CTA UI:
  - `AppHeader` notification panel memakai `BrowserPushCard.svelte` versi compact
  - detail issue menampilkan `BrowserPushCard.svelte` di bawah follow card setelah browser anonim mengikuti issue
  - panel preferensi menampilkan toggle push, tetapi saat browser push belum aktif toggle dapat dinonaktifkan dan user diarahkan ke `BrowserPushCard`
  - khusus iPhone/iOS:
    - jika JEDUG dibuka dari tab Safari biasa, card tidak memakai status `Tidak didukung`
    - UI menampilkan copy bahwa Web Push hanya aktif jika app ditambahkan ke Home Screen lalu dibuka dari ikon app
    - card menampilkan langkah singkat `Share -> Add to Home Screen -> buka ulang dari ikon`
    - flow permission/subscription baru dijalankan setelah app memang berada di mode standalone/Home Screen
  - jika backend mengembalikan `follower_binding_not_found` saat refresh `follower_token`:
    - `BrowserPushCard` masuk state recovery khusus
    - card menjelaskan bahwa browser lokal perlu di-reset lalu consent JEDUG perlu diulang
    - tombol `Reset browser ini` menghapus identitas anonim lokal (`anon_token`, consent flag, follower id/token) lalu reload agar popup consent muncul lagi
  - notification center juga menampilkan panel `Nearby Alerts` dengan UX minimum:
    - tambah lokasi pantauan baru
    - isi otomatis koordinat dari browser jika user mengizinkan
    - input manual latitude/longitude
    - edit label + radius
    - aktif/nonaktifkan watched location
    - hapus watched location
  - Nearby Alerts tetap memakai token notif non-SSE yang sama dengan notifikasi/push ditambah `X-Device-Token`, sehingga tidak menambah auth flow baru.
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
- Route melakukan fetch:
  - `GET /api/v1/stats`
  - `GET /api/v1/stats/regions/options`
  via `lib/api/stats.ts`.
- Query yang dipakai frontend:
  - initial/default: tanpa query agar backend mengembalikan scope default
  - manual filter: `province_id`, `regency_id`
- Struktur section:
  - Filter Wilayah (provinsi + kabupaten/kota)
  - Ringkasan Scope Aktif (card grid)
  - Status Breakdown (card + progress bar)
  - Time Stats (rata-rata umur issue + issue tertua unresolved)
  - Region Leaderboard administratif (list wilayah)
  - Top Issue (card list dengan link ke `/issues/[id]`)
- State wajib:
  - loading
  - error + retry
  - empty
- Default wilayah:
  - backend tetap memberi `filters.active_province_id` + `filters.active_regency_id` fallback agar page punya scope awal yang stabil
  - dropdown manual diisi dari endpoint opsi wilayah terpisah; frontend menyimpan daftar provinsi + regency agar select tidak kosong saat user ingin override scope
  - frontend mencoba geolocation browser + `GET /api/v1/location/label`
  - pencocokan nama wilayah kini menormalkan alias seperti `Daerah Khusus Ibukota Jakarta`, `Kota Administrasi Jakarta Selatan`, kapitalisasi, spasi, dan prefix `Kabupaten/Kota`
  - jika lokasi hanya cocok di level provinsi, page pindah ke scope provinsi dan user diminta memilih kabupaten/kota secara manual
  - jika geolocation gagal/tidak tersedia, page tetap memakai default backend dan user bisa ganti manual
- UX filter wilayah:
  - tombol `Gunakan lokasi saya` tersedia sebagai retry manual tanpa reload halaman
  - dropdown provinsi menampilkan loading/error helper yang jelas, bukan kosong tanpa alasan
  - dropdown kabupaten/kota baru aktif setelah provinsi terpilih
  - memilih provinsi tidak lagi memaksa kabupaten/kota pertama; user bisa melihat scope level provinsi atau lanjut memilih kota/kabupaten
- Contract stats yang dipakai page:
  - `global`: snapshot seluruh issue publik untuk pembanding
  - `summary`: totals yang mengikuti `active_scope`
  - `status` + `time`: juga mengikuti `active_scope`
  - `active_scope.kind/label/is_default`: metadata scope yang dipakai seluruh section utama
- Copy filter dan heading summary kini menegaskan bahwa ringkasan, status, time stats, leaderboard, dan top issue memakai scope aktif yang sama; snapshot global hanya ditampilkan sebagai pembanding kecil saat scope tidak global.
- Region leaderboard tidak lagi memakai identity/key dari `district_name`; frontend merender `regions[].region_id` + `regions[].region_name`, lalu menampilkan konteks parent administratif (`regency/province`) bila ada.
- Top issue card sekarang menampilkan:
  - judul issue (`road_name`, lalu fallback area/wilayah manusiawi bila `road_name` lama hanya berupa label koordinat sintetis)
  - lokasi administratif ringkas dari `district_name`, `regency_name`, `province_name`
  - metrik ringkas (`laporan`, `korban`, `umur issue`)
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
4. Request presign upload dengan `anon_token`.
   - response kini juga membawa `upload_token` + expiry pendek
5. Upload binary.
   - jika upload mode `local`, frontend wajib kirim header `X-Upload-Token`
   - jika upload R2 gagal, fallback ke endpoint local `/api/v1/uploads/file/{object_key}` dengan header `X-Upload-Token` yang sama
6. Submit report dengan metadata media + `upload_token` + `client_request_id`.
   - `client_request_id` dipertahankan stabil sepanjang satu sesi form agar retry submit yang sama bisa di-replay aman oleh backend.

- payload juga mengirim `actor_follower_id` (identity follower anonim dari browser) untuk mencegah self-notify pada event yang dipicu user itu sendiri.
- backend melakukan normalisasi lokasi saat submit (region internal + reverse geocode road fallback).
- frontend tidak mengirim field label lokasi terpisah saat submit; backend tetap authoritative. Label UX di `/lapor` dipakai untuk preview, lalu backend mengulang normalisasi dari koordinat yang sama agar issue publik konsisten.
- UI tetap menampilkan ini sebagai label UX, bukan input wajib user.

Hardening UX submit report:

- sebelum submit, route `/lapor` selalu memastikan bootstrap device anonim selesai (`ensureDeviceBootstrap`)
- bootstrap memakai guard promise bersama + retry ringan untuk mencegah race condition antar inisialisasi layout dan submit
- bila backend mengembalikan indikasi token bootstrap tidak valid (`device not found; bootstrap first`), frontend melakukan refresh bootstrap sekali lalu retry submit otomatis
- bila backend mengembalikan `409 IDEMPOTENCY_CONFLICT`, UI menampilkan pesan eksplisit bahwa key submit lama sudah dipakai untuk payload berbeda dan user perlu memulai submit baru
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
