# JEDUG UI Polish — Panduan & Ringkasan

## Scope dan Source of Truth

Dokumen ini menjadi source of truth untuk:

- guidance UX level halaman
- ringkasan keputusan polish UI lintas halaman
- konteks implementasi visual agar konsisten antar route

Dokumen ini bukan sumber utama token global atau spec detail komponen:

- token global: `design-docs/design-system.md`
- behavior komponen: `design-docs/component-spec.md`

## Apa yang Sudah Dikerjakan

Seluruh frontend JEDUG (Svelte 5 + SvelteKit 2) telah di-polish mengikuti design system yang konsisten dan civic-grade. Tidak ada redesign — hanya penyelarasan warna, tipografi, spacing, dan komponen.

---

## File Design Reference

| File                            | Isi                                                               |
| ------------------------------- | ----------------------------------------------------------------- |
| `design-docs/design-system.md`  | Token lengkap: warna, tipografi, spacing, radius, shadow, z-index |
| `design-docs/component-spec.md` | Spesifikasi per komponen UI                                       |
| `design-docs/guide.md`          | File ini — page-level UX guidance + ringkasan perubahan           |

---

## Komponen yang Di-polish

### Global (`+layout.svelte`)

- Font family: Inter (Google Fonts)
- Background: `#F8FAFC`, text: `#0F172A`, base size: `14px`
- Init toast styling diselaraskan

### AppHeader

- Brand color `#E5484D`, sub-label "Pantau Jalan Rusak"
- Nav link hover: `#FEF2F2` bg tint, padding `6px 12px`
- Sub-label hidden di viewport < 360px

### IssueMap (Marker + Heatmap)

- Severity colors: Ringan `#F6C453`, Sedang `#F97316`, Berat `#DC2626`
- Touch target 36px, dot sizes 14/16/20px
- Selected: `scale(1.5)` + red glow; Fixed: opacity `0.45`
- Heatmap mode memakai ramp kuning -> oranye -> merah tua agar tetap terbaca di basemap light
- Heatmap harus tetap clean, tidak neon, dan menunjukkan pola area tanpa menggantikan marker mode

### IssueBottomSheet

- 20px top radius, drag handle 40×4px `#CBD5E1`
- Desktop: 380px side panel; Mobile: 55vh max
- Status badges color-coded (Open=biru, Verified=hijau, Fixed=abu)
- Buttons min-height 48px, primary `#E5484D`

### PrimaryButton

- Background `#E5484D`, min-height 48px
- Disabled opacity `0.45`, active `scale(0.97)`

### IssueCard

- Severity & status color-coded badges
- Border-radius 16px, font 12px badges

### EmptyState

- Props baru: `ctaHref` + `ctaLabel` untuk tombol CTA opsional
- Padding 48px, button style `#E5484D`

### ErrorState

- Warna `#DC2626`, bg `#FEF2F2`, border `#FECACA`
- Radius 12px

### LoadingState

- Spinner `#E5484D`, track `#E2E8F0`, padding 48px

### ConsentSheet

- Heading 18px/700, body `#64748B`
- Button `#E5484D`, min-height 48px, `scale(0.97)` active

### ImagePicker

- Placeholder bg `#F8FAFC`, border `#E2E8F0` dashed
- Hover `#F1F5F9` + `#CBD5E1` border
- Change label 12px/500

---

## Halaman yang Di-polish

### Landing (`/`)

- Hero diperkuat dengan kicker/trust statement ringan + card treatment lembut
- Hierarki teks diperjelas: brand title tegas, tagline lebih readable, intro lebih terstruktur
- CTA utama/sekunder diselaraskan (radius 12px, min-height ~52px, shadow lembut)
- Tambah CTA `Statistik Jalan Rusak` agar route `/stats` mudah ditemukan dari homepage.

### Peta & Daftar Laporan (`/issues`)

- Map overlay 10px radius, side panel 360px
- Empty state map tidak memakai popup tengah; status area ditampilkan sebagai info badge di kiri atas
- Toolbar, badges, filter semua ke design spec
- Bottom CTA `#E5484D`, 48px min-height
- Guard state transisi list ↔ map untuk menghindari false-empty/flicker saat map remount
- Empty state map hanya tampil setelah fetch viewport valid, bukan saat map baru mount
- Bottom sheet mobile mendukung swipe-down close dengan threshold + snap-back
- Tambahkan segmented toggle `Marker` / `Heatmap` yang tetap nyaman di mobile
- Marker map publik menggunakan clustering:
  - zoom out: marker bergabung ke cluster yang menampilkan count
  - zoom in: cluster pecah otomatis ke marker individual
  - klik cluster: map auto zoom/focus ke area cluster
- Marker individual tetap clickable dan tetap membuka bottom sheet ringkas
- Heatmap mode:
  - menyembunyikan marker individual dan cluster agar peta tidak penuh
  - memakai weight severity-aware dengan bonus korban dan bonus kecil jumlah laporan
  - menampilkan legend intensitas ringkas agar user cepat paham arti warna
- Jika heatmap gagal load, UI harus fallback ke marker mode dan tetap mempertahankan peta aktif
- Jika setup clustering gagal, map fallback ke marker individual layer agar tetap usable
- Saat halaman peta dibuka, geolocate dipicu otomatis sekali pada first load untuk memunculkan blue dot lokasi user jika izin lokasi tersedia.
- Tombol geolocate manual tetap tersedia sebagai retry; map tidak boleh recenter terus-menerus setelah user berinteraksi.

### Statistik Publik (`/stats`)

- Halaman harus mobile-first dan ringan (tanpa grafik berat) agar aman untuk perangkat low-end.
- Hero section memakai brand red `#E5484D` sebagai accent untuk konteks civic storytelling.
- Struktur konten tetap konsisten dan mudah di-scan:
  1. Filter wilayah administratif (`provinsi` + `kabupaten/kota`)
  2. Ringkasan scope aktif (card grid)
  3. Status breakdown card + bar sederhana
  4. Time stats (rata-rata umur issue + issue tertua unresolved)
  5. Region leaderboard list berbasis wilayah administratif
  6. Top issue cards
- User harus bisa membedakan data global vs scoped:
  - card utama mengikuti scope aktif
  - jika UI tetap menampilkan pembanding global, labelnya harus eksplisit dan sekunder
- Default filter harus mencoba memakai lokasi user saat ini.
- Jika geolocation gagal atau belum tersedia, halaman boleh memakai scope default backend asalkan user tetap bisa mengganti wilayah manual.
- Filter wilayah harus tetap usable walau lokasi default tidak match persis:
  - dropdown provinsi terisi dari dataset statistik yang tersedia
  - dropdown kabupaten/kota baru aktif setelah provinsi dipilih
  - tampilkan helper/loading yang jelas saat opsi wilayah sedang diambil
  - sediakan tombol retry ringan `Gunakan lokasi saya`
- Region leaderboard tidak lagi memakai label pseudo-lokasi seperti `Sekitar Jalan ...`; prioritasnya identity wilayah administratif yang stabil, lalu label manusiawi dari wilayah itu.
- Top issue card harus menampilkan konteks lokasi ringkas `kecamatan, kabupaten/kota, provinsi` bila data tersedia.
- Top issue wajib menyediakan link cepat ke detail `/issues/{id}`.
- State wajib:
  - loading
  - error + retry
  - empty data
- Spacing, border, radius, typography tetap mengikuti token utama:
  - card radius 16/12px
  - border `#E2E8F0`
  - text primary `#0F172A`
  - text secondary `#64748B`

### Form Lapor (`/lapor`)

- Title 20px/700, section spacing 24px
- Severity selected `#E5484D`, bg `#F1F5F9`
- Submit `#E5484D`, disabled 0.45, active scale(0.97)
- Semua px units, warna sesuai spec
- Setelah koordinat tersedia, tampilkan label lokasi manusiawi dari lookup wilayah internal
- Label lokasi hanya bersifat konfirmasi UX; koordinat mentah tetap ditampilkan dan tetap jadi acuan submit
- Jika label gagal didapat, user tetap bisa submit report tanpa blocking
- Panel lokasi menampilkan format label primary/secondary agar lebih mudah dibaca di mobile.
- Tambahkan helper text bahwa nama jalan issue akan dilengkapi otomatis saat submit report.
- Submit report harus selalu melewati guard bootstrap device anonim sebelum request report dikirim.
- Jika bootstrap belum siap/gagal, user mendapat copy error manusiawi (bukan pesan backend mentah).

### Header Navigasi Global

- Active state tab `Lapor`, `Peta`, `Statistik` harus langsung sinkron dengan route saat initial render, refresh, dan navigasi client-side.
- Source of truth active state adalah pathname route, bukan state lokal berbasis click.
- Notification center di header harus tetap mobile-first:
  - tiap item memiliki action hapus ringan
  - unread badge harus langsung sinkron saat item dibaca atau dihapus
  - panel tidak boleh terasa berat atau memaksa reload penuh
  - card `Notifikasi Browser` boleh muncul di atas list sebagai CTA ringan, bukan popup agresif
  - CTA ini harus menjelaskan bahwa browser push adalah channel tambahan di atas notifikasi dalam aplikasi
  - khusus iPhone/iOS tab browser biasa, CTA harus menjelaskan syarat Home Screen app sebelum push bisa aktif
  - jika binding follower browser putus, CTA harus menyediakan jalan pulih yang jelas: reset browser lokal lalu consent ulang
  - tambahkan panel `Preferensi Notifikasi` ringan di dropdown yang sama agar user tidak perlu masuk halaman settings baru
  - tambahkan panel `Nearby Alerts` ringan di dropdown yang sama agar user bisa memantau area lokal tanpa follow issue satu per satu
  - preferensi minimum:
    1. master switch semua notifikasi
    2. channel in-app
    3. channel push
    4. event foto baru
    5. event perubahan status
    6. event perubahan tingkat keparahan
    7. event laporan korban baru
    8. event laporan baru di area pantauan
  - jika browser push belum aktif, status harus jelas dan user diarahkan ke CTA aktivasi browser push yang sudah ada
  - Nearby Alerts di notification center harus mobile-first:
    - form tambah lokasi ringkas
    - tombol `Gunakan lokasi saya` opsional
    - input manual tetap tersedia
    - item lokasi bisa diaktif/nonaktifkan dan dihapus tanpa pindah halaman

### Detail Laporan (`/issues/[id]`)

- Layout mobile-first dengan urutan informasi:
  1. hero media utama (full-width mobile, rounded desktop)
  2. badges severity/status/verification
  3. metrik ringkas (laporan/foto/korban/reaksi/update terakhir)
  4. galeri media publik
  5. detail tambahan + catatan publik aman (jika ada)
  6. aktivitas terbaru + CTA share/action
- Shareability built-in:
  - tombol share
  - link cepat WhatsApp/Telegram/Twitter(X)/Facebook
  - metadata SEO/OG/Twitter canonical disiapkan dari SSR data route
  - OG image memakai generator dinamis `/api/og/issues/{id}` dengan komposisi teks issue + background foto (jika ada) atau gradient brand fallback
- Desktop issue detail memakai container lebih lebar dari route publik biasa, dengan card action/share di kolom samping agar halaman terasa lebih lega tanpa mengubah flow `/issues` map-first.
- Fallback states wajib:
  - not found
  - error
  - loading retry
  - fallback media rusak
  - empty gallery
- Tambahkan section `Riwayat Laporan` sebagai timeline vertikal mobile-first:
  - event terbaru di atas
  - marker visual sederhana (`●` + garis vertikal)
  - event minimal: issue dibuat, foto ditambah, severity berubah, korban dilaporkan, status issue berubah
  - jika event > 100, gunakan pagination bertahap (`Muat event lebih lama`)
- Tambahkan card `Ikuti Perkembangan` sebagai CTA follow issue anonim:
  - tampil tanpa login penuh
  - follower count terlihat jelas
  - tombol toggle follow/unfollow inline tanpa reload halaman
  - helper text menjelaskan bahwa browser anonim ini menjadi identitas follow sementara
  - setelah user follow, boleh tampil CTA tambahan `Aktifkan Notifikasi Browser`
  - CTA browser push hanya meminta permission setelah user menekan tombolnya
- Catatan publik di issue detail tidak boleh memakai note mentah bila sudah ada `public_note` yang lebih aman dan ringkas dari API.
- Jika user membuka notifikasi untuk issue yang sedang aktif:
  - halaman tidak melakukan navigasi sia-sia
  - data issue, timeline, dan follow state di-refresh lokal
  - tampilkan feedback ringan seperti `Laporan diperbarui`
- Jika push datang saat tab JEDUG sedang visible:
  - hindari OS notification ganda
  - tab aktif boleh menerima refresh issue lokal via message dari service worker
- Visual tetap mengikuti token:
  - severity colors (kuning-oranye-merah)
  - status/verification badge color-coded
  - card white + border `#E2E8F0`, radius 16px
  - CTA primary `#E5484D`, min-height 48px

---

## Pola Perubahan Umum

| Sebelum              | Sesudah     | Alasan                          |
| -------------------- | ----------- | ------------------------------- |
| `#e53e3e`            | `#E5484D`   | JEDUG Red brand konsisten       |
| `#4a5568`, `#718096` | `#64748B`   | Neutral text secondary          |
| `#a0aec0`            | `#94A3B8`   | Muted text                      |
| `#2d3748`            | `#0F172A`   | Text primary                    |
| `#f7fafc`            | `#F8FAFC`   | Background                      |
| `#edf2f7`            | `#F1F5F9`   | Surface                         |
| `#e2e8f0`            | `#E2E8F0`   | Border                          |
| `rem` units          | `px` units  | Konsistensi, predictable sizing |
| Emoji di CTA         | Text only   | Lebih profesional               |
| Status 1 warna       | Color-coded | Informatif                      |

---

## Apa yang Belum Diubah

- **Flow map publik utama (`/issues`)** — tetap map-first, tidak dirombak.
- **Flow submit report (`/lapor`)** — tidak berubah.
- **Tailwind/CSS framework** — tetap inline scoped styles (tanpa migrasi framework).

## Read This Next

- `design-docs/design-system.md`
- `design-docs/component-spec.md`
- `docs/DESIGN_INDEX.md`

TEST DEPLOY
