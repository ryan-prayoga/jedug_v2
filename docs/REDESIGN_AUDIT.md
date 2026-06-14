# JEDUG Redesign Audit — Index

Dokumen ini meringkas audit menyeluruh sebelum full UI/UX redesign jedug ke arah **minimalist editorial** (warm monochrome, typographic contrast, flat bento, no gradients, no heavy shadows, magazine-grade).

Audit dipecah ke beberapa file karena volume; baca berurutan:

1. [`docs/redesign-audit/00-brief.md`](redesign-audit/00-brief.md) — North-star design brief (taste, tokens target, principles, anti-patterns yang dilarang).
2. [`docs/redesign-audit/01-surfaces.md`](redesign-audit/01-surfaces.md) — Inventaris lengkap route + komponen + form yang harus diredesain.
3. [`docs/redesign-audit/02-tokens-gap.md`](redesign-audit/02-tokens-gap.md) — Gap analysis token (`design-docs/*` vs `app.css` vs kode).
4. [`docs/redesign-audit/03-antipatterns.md`](redesign-audit/03-antipatterns.md) — Hit-list AI-slop & antipattern dengan `path:line + fix`.
5. [`docs/redesign-audit/04-redesign-plan.md`](redesign-audit/04-redesign-plan.md) — Rencana eksekusi (tokens → komponen → halaman → polish), urutan & success criteria.

## TL;DR

- **Taste**: warm off-white, hairline borders, no shadows, no gradients, serif display + sans body, `01/02/03` editorial numbering.
- **Surface scope**: 8 public route + 4 admin route + 16 reusable component. Tidak ada yang aman dari sentuhan.
- **Token gap**: spacing/radius/typography scale ada di doc tapi belum ada di CSS; severity-4 missing; 4 file copy-paste array warna severity.
- **Antipatterns terbesar**: 3-level nested cards, brand-red glow shadows, gradient backdrops di body, color-only severity encoding, motion tanpa `prefers-reduced-motion`.
- **Eksekusi**: tokens dulu (1 file), komponen primitif (`<StatusPill>` `<SeverityPill>` `<MetricTile>` `<Spinner>` `<Card>`), lalu halaman per halaman.

## Sumber Audit

3 explore agent paralel berjalan di sesi lama:

- `bg_782adb5f` — frontend page structure (3m 31s)
- `bg_ee247d70` — design tokens & styling (4m 27s)
- `bg_9eccc9d7` — UI antipatterns & AI-slop (6m 15s)

Output mentah disimpan di `/tmp/jedug_audits/` (gitignored) dan disinkronkan ke folder `docs/redesign-audit/` sebagai source of truth.

## Read This Next

- `docs/redesign-audit/00-brief.md`
- `docs/DESIGN_INDEX.md`
- `design-docs/design-system.md`
