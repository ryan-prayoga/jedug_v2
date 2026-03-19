## Migration Chain

Folder ini menyimpan migration additive/idempotent untuk upgrade dari baseline historis JEDUG v2 ke schema yang saat ini diasumsikan code.

Aturan:

- urutkan file dengan prefix timestamp.
- migration harus aman dijalankan ulang (`IF NOT EXISTS`, `ADD COLUMN IF NOT EXISTS`, dsb).
- jangan edit migration lama kecuali ada bug yang benar-benar membuat migration tidak executable; untuk perubahan baru, tambah file baru.

Fresh bootstrap:

- gunakan `backend/scripts/bootstrap_db.sh fresh`

Upgrade DB lama:

- gunakan `backend/scripts/bootstrap_db.sh upgrade`

Verifikasi:

- gunakan `backend/scripts/verify_schema_governance.sh`
