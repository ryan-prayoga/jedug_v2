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

- warna marker mengikuti severity (dengan override status fixed/archived)
- ukuran dot berbeda per severity
- selected marker:
  - scale up
  - glow effect
  - map `flyTo` ke marker
- klik area map kosong akan clear selected issue

## Bottom Sheet (Ringkasan Issue)

- mobile: sheet dari bawah
- desktop: side panel kanan
- tampilkan severity/status/location/stats/action CTA

Ringkasan issue bottom sheet:

- overlay memakai `position: absolute`
- state visibility tergantung `selectedIssue`
- interaksi ini sensitif terhadap perubahan event click marker/map

## Geolocation Initial Center

- map default center Indonesia (`[110.4, -7.0]`, zoom 7)
- setelah map load, coba geolocate user sekali (`didAutoCenter` guard)
- jika sukses -> `flyTo` ke user location (zoom 15)
- jika gagal -> tetap di default center

Di halaman `/lapor`:

- geolocation high accuracy dengan fallback mode lebih longgar
- jika tetap gagal, user bisa input koordinat manual

## Hal Sensitif yang Jangan Dirusak

- contract field koordinat dari backend: `latitude`, `longitude`
- alur emit bbox + debounce fetch
- marker select -> bottom sheet -> link detail
- fallback mode list saat map component gagal load
- sync status/severity UI terhadap enum backend

## Current Implementation

- map-first UX berjalan baik untuk browse issue.
- fallback ke mode list sudah ada saat map gagal.

## Known Mismatch

- Perbedaan mapping label status antar komponen bisa membuat label tidak konsisten.
- beberapa perubahan kecil pada DOM/event marker bisa memicu regressi selected marker vs sheet behavior.

## Read This Next

- `docs/FRONTEND.md`
- `docs/BACKEND.md`
- `docs/SCHEMA.md`
