# JEDUG Repository Readme

Ini adalah entrypoint ringkas dokumentasi di root project.

## Mulai dari Sini

- Agent/kontributor teknis: `AGENTS.md`
- Dokumentasi teknis terpusat: `docs/`
- Dokumentasi desain (source of truth UI/UX): `design-docs/`

## Urutan Baca Disarankan

1. `AGENTS.md`
2. `docs/PROJECT_OVERVIEW.md`
3. `docs/ARCHITECTURE.md`
4. `docs/SCHEMA.md`
5. `docs/BACKEND.md` atau `docs/FRONTEND.md` sesuai area kerja
6. `docs/DESIGN_INDEX.md` lalu `design-docs/*`

## Catatan

- Jangan duplikasi dokumentasi desain di `docs/`.
- Jika ada perubahan signifikan, wajib update:
  - dokumen terkait di `docs/`
  - `docs/CHANGELOG_FOR_AGENTS.md`
  - `AGENTS.md` jika onboarding/flow kerja berubah
  - `design-docs/*` jika perubahan menyentuh visual, behavior komponen, atau page-level UX
