# Business Requirements Document (BRD)
## Aplikasi Game Edukasi Matematika (Codename: "MathQuest")

**Versi:** 0.1 (Draft)
**Tanggal:** 13 September 2026
**Author:** Yaaqin

---

## 1. Ringkasan Eksekutif

MathQuest adalah platform latihan matematika berbasis gamifikasi yang dirancang untuk mendampingi pembelajaran numerasi dari jenjang SD hingga topik "umum" (dewasa/hobi). Dokumen ini menjelaskan kebutuhan bisnis di balik pembangunan produk, termasuk tujuan, ruang lingkup bisnis, model monetisasi masa depan, serta risiko dan asumsi.

## 2. Latar Belakang Bisnis

Kebutuhan akan alat bantu latihan matematika yang terstruktur, adiktif secara positif (dengan mekanik game seperti streak & nyawa), dan bisa dipakai mandiri oleh siswa dari berbagai jenjang, mendorong pembuatan produk ini. Produk dibangun dari inisiatif personal (bootstrap), dikembangkan sendiri (owner merangkap product owner & developer), dengan rencana ekspansi bertahap.

## 3. Tujuan Bisnis

1. Memvalidasi apakah model gamifikasi (streak, nyawa, batch/level) efektif meningkatkan engagement latihan matematika.
2. Membangun basis pengguna terdaftar (akun) sebagai aset jangka panjang sebelum monetisasi diaktifkan.
3. Menyiapkan arsitektur yang dapat diperluas ke mata pelajaran lain di luar matematika.
4. Menyiapkan fondasi model bisnis subscription untuk monetisasi di masa depan, tanpa mengorbankan pengalaman gratis di fase awal.

## 4. Pemangku Kepentingan (Stakeholders)

| Peran | Deskripsi |
|---|---|
| Product Owner / Developer | Yaaqin — merangkap sebagai pemilik produk, desainer sistem, dan pengembang (fullstack) |
| End User – Pelajar | Siswa SD/SMP/SMK, mahasiswa, dan pengguna umum |
| Admin Konten | Pengelola bank soal, kategori, dan level (bisa Yaaqin sendiri di awal, atau tim kontributor soal di masa depan) |

## 5. Kebutuhan Bisnis (Business Requirements)

| ID | Kebutuhan Bisnis |
|---|---|
| BR-01 | Sistem harus mendukung registrasi akun via Google SSO maupun username/email + password, agar barrier masuk rendah bagi berbagai jenis pengguna |
| BR-02 | Progres belajar setiap akun harus tersimpan permanen dan bisa diakses lintas sesi/perangkat |
| BR-03 | Konten soal (bank soal, kategori, tier) harus bisa dikelola tanpa bantuan developer (lewat dashboard admin), agar operasional konten scalable |
| BR-04 | Struktur kurikulum harus mencerminkan jenjang pendidikan riil di Indonesia (SD, SMP, SMK) plus jenjang lanjutan (kuliah) dan kategori bebas (umum), agar relevan dengan kebutuhan belajar formal maupun non-formal |
| BR-05 | Mekanisme progres berjenjang (batch + exam pembuka batch berikutnya) diperlukan untuk memastikan penguasaan materi sebelum lanjut ke level berikutnya |
| BR-06 | Sistem harus punya mekanik retensi (streak, nyawa) untuk mendorong penggunaan berkelanjutan, mengikuti pola aplikasi edukasi populer di pasar |
| BR-07 | Arsitektur data harus future-proof untuk model subscription (freemium), meskipun seluruh fitur gratis di fase awal — supaya tidak perlu migrasi skema besar saat monetisasi diaktifkan |
| BR-08 | Sistem harus tahan terhadap kondisi jaringan buruk saat sesi latihan berlangsung (soal di-fetch di awal sesi) agar pengalaman belajar tidak terganggu secara teknis |
| BR-09 | Produk harus mendukung dark mode dan multi-bahasa agar dapat menjangkau basis pengguna yang lebih luas dan nyaman dipakai dalam berbagai kondisi |
| BR-10 | Roadmap harus memungkinkan penambahan mode multiplayer di fase kedua tanpa merombak fondasi yang dibangun di fase pertama |

## 6. Model Monetisasi (Rencana Jangka Panjang)

- **Fase 1 & 2:** Seluruh fitur gratis, tanpa gateway pembayaran aktif.
- **Fase lanjutan (~fase 10):** Model **freemium/subscription** untuk membuka fitur tertentu (misal: kuota nyawa lebih besar, akses konten premium, mode multiplayer lanjutan, dsb — detail fitur premium belum difinalisasi).
- Skema database subscription/entitlement disiapkan dari fase 1 agar transisi ke model berbayar tidak memerlukan migrasi data besar.

## 7. Ruang Lingkup Bisnis

**Termasuk:**
- Produk edukasi matematika lintas jenjang pendidikan Indonesia + kategori umum non-formal.
- Model akun personal (bukan institusi/sekolah) di fase awal.

**Tidak termasuk (saat ini):**
- Kemitraan resmi dengan institusi pendidikan (sekolah/kampus) — dapat dipertimbangkan di fase lanjut.
- Konten mata pelajaran non-matematika (disiapkan strukturnya, belum diisi).
- Model bisnis B2B (misal lisensi ke sekolah).

## 8. Risiko & Asumsi

| Jenis | Deskripsi |
|---|---|
| Risiko | Kualitas & kalibrasi kesulitan soal bergantung penuh pada proses input manual admin — perlu proses kurasi konten yang konsisten |
| Risiko | Tanpa validasi pasar awal, model gamifikasi (nyawa, streak) belum tentu cocok dengan preferensi semua segmen usia (misal siswa SD vs mahasiswa) |
| Risiko | Rencana subscription di fase jauh (~10) berarti model bisnis belum tervalidasi secara finansial dalam waktu dekat |
| Asumsi | Pengguna memiliki akses internet yang cukup stabil untuk fetch soal di awal sesi (meski tidak perlu stabil sepanjang sesi berlangsung) |
| Asumsi | Owner akan berperan sebagai admin konten awal sebelum ada kontributor tambahan |

## 9. Kriteria Keberhasilan (Business Success Criteria)

- Berhasil meluncurkan Fase 1 dengan struktur kurikulum SD lengkap sebagai pilot sebelum memperluas ke jenjang lain.
- Retensi harian pengguna terekam melalui fitur streak sebagai indikator awal product-market fit.
- Dashboard admin operasional digunakan untuk menambah bank soal tanpa keterlibatan langsung developer di setiap perubahan konten.