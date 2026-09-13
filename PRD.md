# Product Requirements Document (PRD)
## Aplikasi Game Edukasi Matematika (Codename: "MathQuest")

**Versi:** 0.1 (Draft)
**Tanggal:** 13 September 2026
**Author:** Yaaqin

---

## 1. Latar Belakang & Visi

Aplikasi ini dibuat untuk melatih kemampuan berpikir dasar matematika melalui pendekatan gamifikasi (game-based learning), mirip pola Duolingo tapi untuk numerasi/matematika. Fokus build awal adalah **matematika**, dengan arsitektur yang dirancang cukup generik agar di fase lanjut bisa diperluas ke mata pelajaran lain (fisika, kimia, dll) tanpa rombak besar di skema data maupun sistem gameplay.

Aplikasi menyasar pengguna dari jenjang SD sampai jenjang "umum" (dewasa/topik non-formal seperti aritmatika kripto), dengan progres belajar yang tersimpan per akun.

## 2. Tujuan Produk

- Menyediakan platform latihan matematika terstruktur per jenjang pendidikan (SD → SMP → SMK → Kuliah → Umum).
- Membuat pengalaman belajar terasa seperti game: ada progres, streak, nyawa (lives), dan tantangan berjenjang (batch & exam).
- Menyimpan progres tiap user secara individual sehingga bisa dilanjutkan kapan saja.
- Menyediakan dashboard admin untuk mengelola bank soal, kategori, dan level tanpa perlu deploy ulang aplikasi.
- Menyiapkan fondasi teknis yang scalable untuk fitur multiplayer (fase 2) dan subscription (fase jauh ke depan).

## 3. Target Pengguna

| Segmen | Deskripsi |
|---|---|
| Siswa SD | Kelas 1–6, materi dasar: penjumlahan, pengurangan, perkalian, pembagian |
| Siswa SMP | Materi aljabar dasar, pecahan, geometri dasar, dll |
| Siswa SMK/SMA | Materi lanjutan sesuai kurikulum SMK/SMA |
| Mahasiswa | Materi aljabar lanjut & kalkulus |
| Umum | Dewasa/hobi, topik non-formal (mis. "crypto arithmetic"), tidak berbasis kurikulum sekolah |

## 4. Ruang Lingkup (Scope)

### Fase 1 — Fondasi & Registrasi
- Registrasi & login: Google SSO, atau username/email + password.
- Akun SSO tetap bisa menambahkan password secara manual di kemudian hari (linked auth method).
- Struktur dasar kurikulum: Tier → Batch → Challenge → Exam (khusus tier "Umum": Tier → Kategori → Challenge, tanpa batch).
- Dashboard admin dasar: CRUD tier, batch/kategori, challenge, dan bank soal.
- Gameplay single-player: user mengerjakan challenge, sistem nyawa, streak.
- Dark mode & multi-bahasa (i18n) di UI.

### Fase 2 — Multiplayer
- Mode bermain bersama (real-time atau async, didetailkan di FSD/desain terpisah saat fase ini mulai digarap).

### Fase Lanjutan (belum digarap sekarang, hanya disiapkan skemanya)
- Subscription/monetisasi (kira-kira fase ~10) — skema DB disiapkan dari awal, tapi semua fitur gratis dulu di fase 1 & 2.

### Di Luar Scope (saat ini)
- Mata pelajaran selain matematika (disiapkan strukturnya, belum diisi kontennya).
- Sistem pembayaran aktif (skema disiapkan, gateway belum diimplementasi).
- Native mobile app (asumsi awal: web app responsive dulu).

## 5. Struktur Kurikulum & Gameplay

### 5.1 Hierarki Konten
```
Tier (SD / SMP / SMK / Kuliah / Umum)
 └─ Batch (khusus non-Umum, misal "Batch 1 – Kelas 1")
     ├─ Challenge 1..N
     └─ Exam (membuka Batch berikutnya)
 └─ (Umum) Kategori (misal "Crypto Arithmetic") — tanpa batch
     └─ Challenge 1..N
```

- Tiap **Batch** berisi beberapa Challenge biasa + 1 Exam sebagai gerbang menuju batch berikutnya.
- Tier **Umum** tidak memakai konsep batch berjenjang; kontennya dikelompokkan per **Kategori** yang berdiri sendiri (tidak saling mengunci).

### 5.2 Bank Soal & Challenge
- Setiap Challenge terhubung ke sebuah **bank soal** yang bisa berisi lebih banyak soal daripada yang ditampilkan.
- Contoh: Challenge "SD No. 5" mewajibkan user menjawab 20 soal, tapi bank soalnya bisa berisi >20 soal — soal yang keluar diambil (dan diacak) dari bank tersebut.
- Soal & urutan soal diacak setiap kali user mengambil challenge yang sama.
- Pilihan jawaban (untuk tipe pilihan ganda) juga diacak posisinya, dan direpresentasikan sebagai **nilai langsung** (bukan label A/B/C/D).
- Jumlah pilihan jawaban berbeda per tier:
  - SD: 3 pilihan
  - SMP: 4 pilihan
  - SMK & Umum: 5 pilihan
- Selain pilihan ganda, sistem juga mendukung tipe soal **esai/isian** di mana user mengetik langsung jawaban numeriknya.
- **Model "Gacha Fetch":** seluruh soal untuk satu sesi challenge di-fetch sekaligus di awal sebelum challenge dimulai, supaya user tidak terputus di tengah jalan akibat koneksi buruk.

### 5.3 Gamifikasi
- **Streak:** terhitung dari aktivitas harian user (menyelesaikan minimal satu challenge per hari mempertahankan streak).
- **Sistem Nyawa (Lives):**
  - Reset otomatis: 3 nyawa per hari.
  - Batas maksimum: 15 nyawa per minggu (agar tidak menumpuk tak terbatas).
  - User bisa mereset nyawa secara manual lewat dashboard mereka sendiri (mekanisme reset manual ini perlu aturan — lihat catatan di FSD, misal ada batas pemakaian atau terhubung ke fitur premium di masa depan).

## 6. Kebutuhan Non-Fungsional

- **Dark mode**: seluruh UI mendukung light/dark theme.
- **Multi-bahasa (i18n)**: minimal Bahasa Indonesia & English di fase awal, arsitektur i18n harus mudah menambah bahasa baru.
- **Resiliensi jaringan**: pola fetch-di-awal (gacha) supaya gameplay tidak terganggu koneksi lemah setelah challenge dimulai.
- **Skalabilitas data soal**: admin harus bisa menambah ratusan/ribuan soal tanpa masalah performa saat proses randomisasi & pengambilan bank soal.
- **Auditability progres**: setiap attempt/percobaan user tersimpan agar bisa dianalisis (skor, waktu, jumlah nyawa terpakai, dsb).

## 7. Metrik Keberhasilan (Awal)

- Jumlah user teregistrasi & retensi harian (streak aktif).
- Completion rate per batch/tier.
- Rata-rata skor per challenge (indikator kalibrasi kesulitan soal).
- Jumlah soal aktif di bank soal per tier/kategori.

## 8. Asumsi & Batasan

- Tahap awal aplikasi berbasis web (Next.js), belum tentu native mobile.
- Backend Go + PostgreSQL sudah ditetapkan sebagai stack teknis (lihat FSD untuk detail arsitektur).
- Semua fitur gratis di fase 1 & 2; skema subscription hanya disiapkan di database, belum ada enforcement/gateway pembayaran.
- Mode multiplayer belum didefinisikan detail mekanismenya (real-time vs async) — akan dirinci saat fase 2 dimulai.

## 9. Roadmap Ringkas

| Fase | Fokus |
|---|---|
| 1 | Registrasi, struktur kurikulum, dashboard admin, gameplay single-player, streak, lives, dark mode, i18n |
| 2 | Multiplayer |
| ~10 (jauh ke depan) | Subscription & monetisasi aktif |