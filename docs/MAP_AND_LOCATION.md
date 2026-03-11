# Map and Location Guide

## Teknologi

- Map engine: MapLibre GL
- Basemap: Carto Positron style
- Halaman utama map: `/issues`

## BBox / Viewport Fetching

Flow data peta:

1. `IssueMap` emit bbox saat `load` dan `moveend`.
2. Parent page memanggil `fetchIssuesByBBox`.
3. Request didebounce (default 300ms).
4. Fetch issue pakai query `bbox=minLng,minLat,maxLng,maxLat`.
5. BBox sama (rounded 5 digit) tidak difetch ulang.

Catatan penting:

- backend memakai filter `is_hidden=false` dan exclude status `rejected/merged` untuk publik.
- limit default map fetch saat ini 100 issue/viewport.

## Marker Behavior

- marker publik dirender dari satu `GeoJSON source`:
  - `cluster circles`
  - `cluster count labels`
  - `unclustered hit-area`
  - `unclustered marker dots`
  - `selected glow/core`
- warna marker individual mengikuti severity (dengan override status fixed/archived)
- ukuran dot individual berbeda per severity
- cluster count dibuat tetap terbaca di mobile (symbol layer dengan halo)
- klik cluster melakukan zoom/focus ke cluster (`getClusterExpansionZoom` + `easeTo`)
- selected marker:
  - marker base disembunyikan via filter
  - selected glow + core layer ditampilkan
  - map `flyTo` ke marker
- klik area map kosong akan clear selected issue (guard terhadap klik cluster/marker agar tidak bentrok event)
- fallback safety:
  - jika setup clustering gagal, map fallback ke layer marker individual tanpa cluster
  - flow klik marker + bottom sheet tetap berjalan

## Bottom Sheet (Ringkasan Issue)

- mobile: sheet dari bawah
- desktop: side panel kanan
- tampilkan severity/status/location/stats/action CTA

Ringkasan issue bottom sheet:

- overlay memakai `position: absolute`
- state visibility tergantung `selectedIssue`
- interaksi ini sensitif terhadap perubahan event click marker/map
- mobile sheet bisa di-drag/swipe ke bawah untuk menutup:
  - drag melewati threshold -> close
  - drag pendek -> snap kembali ke posisi awal
  - drag diprioritaskan dari handle dan tidak bentrok dengan scroll internal konten

## Geolocation Initial Center

- map default center Indonesia (`[110.4, -7.0]`, zoom 7)
- setelah map load, coba geolocate user sekali (`didAutoCenter` guard)
- jika sukses -> `flyTo` ke user location (zoom 15)
- jika gagal -> tetap di default center

Di halaman `/lapor`:

- geolocation high accuracy dengan fallback mode lebih longgar
- jika tetap gagal, user bisa input koordinat manual
- setelah koordinat tersedia, frontend memanggil `GET /api/v1/location/label` untuk label lokasi manusiawi
  - trigger saat lokasi awal berhasil didapat
  - trigger saat koordinat manual dipilih eksplisit
  - tidak dipanggil terus-menerus saat user mengetik
  - ada cache in-memory (key koordinat rounded) untuk menekan request berulang
- jika label gagal didapat, koordinat mentah tetap dipakai dan submit report tidak diblokir

## Hal Sensitif yang Jangan Dirusak

- contract field koordinat dari backend: `latitude`, `longitude`
- alur emit bbox + debounce fetch
- marker select -> bottom sheet -> link detail
- fallback mode list saat map component gagal load
- sync status/severity UI terhadap enum backend

## Current Implementation

- map-first UX berjalan baik untuk browse issue.
- fallback ke mode list sudah ada saat map gagal.
- transisi list ↔ map memakai guard untuk mencegah false-empty/flicker saat mount ulang map.
- status jumlah titik + "tidak ada laporan" ditampilkan via badge info top-left, bukan popup tengah.

## Known Mismatch

- Perbedaan mapping label status antar komponen bisa membuat label tidak konsisten.
- beberapa perubahan kecil pada DOM/event marker bisa memicu regressi selected marker vs sheet behavior.

## Read This Next

- `docs/FRONTEND.md`
- `docs/BACKEND.md`
- `docs/SCHEMA.md`
