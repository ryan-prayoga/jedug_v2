# Design Documentation Index

Dokumen ini menghubungkan dokumentasi desain (`design-docs/`) dengan dokumentasi teknis (`docs/`), agar tidak berjalan terpisah.

## Peran Folder `design-docs/`

`design-docs/` adalah source of truth desain JEDUG untuk:

- visual system (warna, tipografi, spacing, layer)
- aturan perilaku komponen UI
- guidance UX level halaman

## Peta Dokumen Desain

### `design-docs/design-system.md`

- Fungsi utama: design tokens dan visual primitives.
- Source of truth untuk:
  - visual style
  - color/token/radius/shadow/z-index

### `design-docs/component-spec.md`

- Fungsi utama: behavior dan spesifikasi komponen UI.
- Source of truth untuk:
  - component behavior
  - state komponen, layout komponen, dan interaction pattern

### `design-docs/guide.md`

- Fungsi utama: guidance level halaman dan ringkasan implementasi UI.
- Source of truth untuk:
  - page-level UX
  - ringkasan arah polish lintas halaman

## Aturan Prioritas Jika Terjadi Konflik

1. Token visual -> `design-system.md`
2. Perilaku komponen -> `component-spec.md`
3. Alur UX per halaman -> `guide.md`
4. Jika konflik masih ada, update ketiga dokumen agar konsisten lalu catat di `docs/CHANGELOG_FOR_AGENTS.md`.

## Keterkaitan Dengan Docs Teknis

- `docs/FRONTEND.md` menjelaskan implementasi aktual route/komponen.
- `docs/MAP_AND_LOCATION.md` menjelaskan behavior map yang harus konsisten dengan spesifikasi desain.
- `AGENTS.md` mengharuskan agent membaca dokumen desain sebelum perubahan UI.

## Sinkronisasi Wajib

Jika ada perubahan:

- design token -> update `design-system.md`
- behavior komponen -> update `component-spec.md`
- UX level halaman -> update `guide.md`
- dampak teknis frontend/API -> update dokumen di `docs/` terkait

## Read This Next

- `design-docs/design-system.md`
- `design-docs/component-spec.md`
- `design-docs/guide.md`
- `docs/FRONTEND.md`
