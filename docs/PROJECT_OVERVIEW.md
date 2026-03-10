# JEDUG Project Overview

## Ringkasan Produk

JEDUG adalah platform pelaporan jalan rusak berbasis partisipasi publik yang berfokus pada:

- pelaporan cepat dari perangkat mobile
- pemetaan issue secara publik
- akumulasi bukti (foto + submission)
- moderasi terpusat untuk menjaga kualitas data

## Visi

Membuat data jalan rusak lebih terbuka, terukur, dan sulit diabaikan melalui pelaporan publik yang mudah.

## Tujuan Utama

- menurunkan friksi warga saat melapor
- mengubah laporan ad-hoc menjadi issue terstruktur
- memberikan visibilitas publik via peta dan detail issue
- memberi admin alat moderasi dan hardening anti-spam

## Masalah yang Diselesaikan

- laporan tersebar di kanal informal dan tidak terstruktur
- minim bukti terpusat (foto, lokasi, riwayat laporan)
- sulit memonitor issue berulang pada titik yang sama
- minim kontrol kualitas data publik

## Positioning Produk

JEDUG diposisikan sebagai civic reporting platform yang:

- `mobile-first` untuk submit cepat
- `map-first` untuk observasi area
- `anonymous-first` agar adopsi awal rendah friksi
- tetap siap evolusi ke akun user/login bila diperlukan

## Ruang Lingkup Implementasi Saat Ini

- anonymous device bootstrap + consent
- upload media local/R2
- submit report -> auto-group ke issue terdekat (10m)
- list/detail issue publik
- admin moderation: hide/unhide/fix/reject issue, ban device
- flag issue oleh komunitas + auto-hide threshold

## Batasan Implementasi Saat Ini

- login user/oauth belum dipakai di alur aplikasi aktif
- issue reactions/submission flags/issue daily stats belum dipakai penuh di service API
- beberapa field schema belum dikelola penuh (contoh: `resolved_at`, `issue_status_history`)

## Read This Next

- `docs/ARCHITECTURE.md`
- `docs/SCHEMA.md`
- `docs/BACKEND.md`
- `docs/FRONTEND.md`
