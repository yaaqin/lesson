# Functional Specification Document (FSD)
## Aplikasi Game Edukasi Matematika (Codename: "MathQuest")

**Versi:** 0.2 (Draft)
**Tanggal:** 13 September 2026
**Author:** Yaaqin
**Stack:** Next.js (App Router) ×2 app · React Native (mobile, fase berikutnya) · Go (backend service) · PostgreSQL
**Update di versi ini:** menambahkan modul Organisasi/Sekolah (grouping kelas, undangan murid, ujian buatan guru) dan struktur monorepo multi-app.

---

## 1. Arsitektur Umum

```
┌────────────┐      REST/JSON       ┌────────────┐      SQL       ┌──────────────┐
│  Next.js   │ ───────────────────▶ │  Go API     │ ─────────────▶ │  PostgreSQL   │
│  (Web App) │ ◀─────────────────── │  Service    │ ◀───────────── │  (primary DB) │
└────────────┘                      └────────────┘                └──────────────┘
      │                                    │
      │ Google OAuth redirect              │ Session/JWT
      ▼                                    ▼
 Google Identity                     Redis (opsional, untuk cache
                                      bank soal aktif per challenge
                                      & rate limit lives)
```

**Catatan teknis:**
- Autentikasi menggunakan kombinasi **JWT (access token)** + **refresh token** tersimpan di HTTP-only cookie.
- Google SSO memakai OAuth2 Authorization Code flow; akun yang login pertama kali via Google otomatis dibuatkan record `users` dengan `auth_provider = google`, dan bisa menambahkan password lokal belakangan (linked credential). Flow yang sama juga dipakai untuk murid yang menerima undangan organisasi (lihat 3.2).
- Redis bersifat opsional untuk fase awal (bisa langsung ke Postgres dulu), tapi direkomendasikan untuk cache bank soal yang sering diakses dan untuk counter lives supaya tidak membebani Postgres dengan write kecil berulang.

### 1.1 Struktur Aplikasi (Monorepo)

Semua kode tinggal dalam satu repo, dipecah per app/service agar deployment & scaling bisa independen:

```
lesson/
├─ apps/
│  ├─ web/          # Next.js — Lesson App, diakses end-user (pelajar umum & murid organisasi)
│  ├─ dashboard/    # Next.js — 1 app, 2 dashboard (role-based routing di dalamnya)
│  │                #   /org/*   -> Dashboard Organisasi (owner/admin/teacher)
│  │                #   /admin/* -> Dashboard Admin Platform (admin/superadmin)
│  └─ mobile/       # React Native — versi mobile Lesson App (fase berikutnya, belum digarap)
├─ services/
│  └─ api/          # Go — backend service, melayani ketiga app di atas
└─ FSD.md
```

**Kenapa dashboard digabung jadi satu app, bukan dua:**
- Org admin & platform admin sama-sama "pengelola", banyak komponen UI (tabel, form CRUD, chart statistik) yang reusable.
- Pemisahan akses cukup lewat middleware role-based routing (bukan lewat repo/app terpisah): user dengan `organization_members.role` (owner/admin/teacher) masuk ke `/org/**`, user dengan `users.role` = admin/superadmin masuk ke `/admin/**`. Seseorang yang kebetulan platform-admin sekaligus punya organisasi bisa switch antar dua area tanpa app/login terpisah.
- Kalau nanti terasa terlalu besar/berat untuk dipisah tim, baru dipecah jadi app sendiri — tidak masalah dilakukan belakangan karena batasnya sudah jelas di level route group.

`apps/web` (Lesson App) tetap satu-satunya app yang diakses pelajar individu maupun murid organisasi untuk main/mengerjakan soal — organisasi tidak punya app terpisah untuk murid, mereka login di app yang sama dan konten organisasi (ujian buatan guru) muncul sebagai bagian dari pengalaman mereka di sana.

## 2. Peran Pengguna (User Roles)

Ada dua lapis role yang **independen satu sama lain**: role platform (global, melekat di `users.role`) dan role organisasi (khusus di dalam satu organisasi, melekat di `organization_members.role`). Satu akun bisa punya kombinasi apa saja, misalnya seorang `student` platform yang juga jadi `teacher` di sebuah sekolah.

**Role Platform (global, `users.role`):**

| Role | Deskripsi | Akses |
|---|---|---|
| `student` | Pengguna umum yang berlatih | Registrasi, main challenge, lihat progres sendiri, reset nyawa sendiri, join organisasi (jika diundang) |
| `admin` | Pengelola konten platform | CRUD tier, batch/kategori, challenge, bank soal global; lihat statistik agregat; kelola daftar organisasi |
| `superadmin` (opsional) | Pengelola admin lain | Manajemen role admin, konfigurasi sistem |

**Role Organisasi (scoped per organisasi, `organization_members.role`):**

| Role | Deskripsi | Akses |
|---|---|---|
| `owner` | Pendaftar pertama organisasi (mis. kepala sekolah / pemilik lembaga) | Full akses ke organisasi, termasuk hapus organisasi, transfer ownership, kelola admin/guru lain |
| `admin` | Staf pengelola organisasi | Kelola grouping/kelas, undang guru & murid, kelola ujian — tidak bisa hapus organisasi |
| `teacher` | Guru | Buat grouping kelas sendiri, buat/kelola ujian (custom soal atau ambil dari bank soal publik), undang murid ke kelasnya, lihat hasil ujian murid |
| `student` | Murid yang join lewat undangan | Akses ujian yang di-assign ke kelasnya, lihat hasil sendiri |

## 3. Modul Fungsional

### 3.1 Autentikasi & Registrasi

**Fitur:**
- Register via email/username + password.
- Register/login via Google SSO.
- **Register sebagai Organisasi** — form terpisah di halaman auth ("Daftar sebagai Sekolah/Organisasi"): input nama organisasi + data akun pemilik (email/password atau Google). Satu submit membuat `users` (jika belum ada) + `organizations` + `organization_members` (role `owner`, status `active`) sekaligus.
- Akun SSO dapat menambahkan password lokal di halaman "Keamanan Akun" (set password flow, bukan login).
- Login, logout, refresh token, forgot/reset password (untuk akun berbasis password).

**Aturan bisnis:**
- Email harus unik di seluruh sistem, terlepas dari metode registrasi (username/password vs Google), untuk mencegah duplikasi akun dengan email sama.
- Jika user register dengan Google menggunakan email yang sama dengan akun password yang sudah ada → tawarkan flow "link akun" (bukan bikin akun baru).
- Password di-hash dengan bcrypt/argon2 (jangan simpan plaintext, jangan pakai algoritma custom).
- Organisasi tidak punya kredensial login sendiri — akses organisasi selalu lewat akun `users` yang sudah punya membership (`organization_members`) di organisasi tersebut.

### 3.2 Organisasi & Grouping (Sekolah, Kelas, Undangan)

**Konsep:**
- **Organisasi** = entitas pendaftar (mis. sekolah), dibuat lewat flow "Register sebagai Organisasi" di 3.1.
- **Grouping/Kelas** (`groups`) = pengelompokan di dalam organisasi, mendukung hierarki 2 level lewat `parent_group_id` — contoh: "Kelas 10" (level/tingkat) → "Kelas 10 A", "Kelas 10 B" (kelas spesifik). Guru bisa bikin grouping sendiri sesuai kebutuhan (tidak harus by tingkat, bisa juga misal per mata pelajaran/angkatan).
- **Ujian Organisasi** (`organization_exams`) = ujian yang dibuat guru, di-assign ke satu grouping/kelas (atau ke seluruh organisasi jika `group_id` kosong).

**Alur undang murid (invite via email, join via Google):**
1. Guru/admin organisasi input email murid + pilih kelas tujuan dari dashboard organisasi.
2. Sistem membuat `organization_members` (status `invited`, `user_id` masih kosong jika murid belum pernah daftar) dan `organization_invites` (token unik + `expires_at`).
3. Sistem mengirim email berisi link undangan (`/invite/:token`).
4. Murid klik link → diarahkan login/register via Google.
5. Backend mencocokkan email dari profil Google dengan `invited_email` pada undangan:
   - **Cocok** → isi `user_id` di `organization_members`, set `status = active`, `joined_at = now()`, tandai `organization_invites.accepted_at`, lalu masukkan ke `group_members` kelas tujuan.
   - **Tidak cocok** → tolak dengan pesan error (mencegah orang lain mengklaim undangan orang lain).
6. Murid otomatis melihat ujian organisasi yang di-assign ke kelasnya saat login ke Lesson App (`apps/web`), berdampingan dengan konten tier/challenge reguler.

**Bank soal untuk guru:**
- Guru bisa membuat soal sendiri, tersimpan privat di bawah organisasinya (tidak terlihat organisasi lain, tidak masuk bank soal publik).
- Guru juga bisa **mengambil soal dari bank soal publik platform** (soal yang dibuat admin platform di 3.4) untuk dipakai di ujian buatannya — jadi satu ujian organisasi bisa berisi campuran soal privat organisasi + soal dari bank publik.
- Soal privat organisasi **tidak otomatis** masuk ke bank soal publik; promosi soal privat → publik (jika dibutuhkan suatu saat) adalah keputusan admin platform, bukan otomatis (lihat Open Decisions #9).

**Aturan bisnis:**
- Satu email undangan hanya boleh punya satu undangan `pending` aktif per organisasi (undangan lama otomatis invalid kalau dikirim ulang).
- `owner` tidak bisa dihapus dari organisasi tanpa transfer ownership ke member `admin`/`teacher` lain dulu.
- Murid bisa jadi anggota lebih dari satu organisasi sekaligus (lihat Open Decisions #7).

### 3.3 Struktur Kurikulum

**Entity hierarki:**
```
tiers (SD, SMP, SMK, Kuliah, Umum)
 └─ batches (khusus tier non-"Umum")
     └─ challenges (termasuk flag is_exam)
 └─ categories (khusus tier "Umum")
     └─ challenges
```

**Aturan bisnis:**
- Tier `umum` tidak memiliki `batches`; challenge-nya menempel langsung ke `categories`.
- Tier selain `umum` wajib mengelompokkan challenge ke dalam `batches`.
- Setiap batch memiliki tepat satu challenge dengan `is_exam = true` sebagai syarat membuka batch berikutnya (`unlock_next_batch = true` jika lulus dengan skor minimum tertentu, mis. `pass_threshold`).
- Urutan batch dan challenge ditentukan oleh kolom `order_index`.
- Struktur ini adalah kurikulum **platform** (global, dikelola admin platform) — terpisah dari `organization_exams` di 3.2 yang scoped ke satu organisasi/kelas saja.

### 3.4 Bank Soal & Manajemen Konten (Admin Dashboard)

**Fitur admin:**
- CRUD Tier, Batch/Kategori, Challenge.
- CRUD Soal (Question) di dalam bank soal per Challenge.
- Set `question_count_required` per challenge (jumlah soal yang harus dijawab per sesi, misal 20).
- Set jumlah opsi jawaban default per tier (3/4/5), dengan kemungkinan override per soal jika diperlukan.
- Preview soal (simulasi tampilan ke user) sebelum publish.
- Toggle status soal: `draft` / `published` / `archived` (soal draft tidak ikut ke-random ke sesi user).

**Tipe Soal:**
1. **Multiple Choice** — opsi jawaban berupa nilai (bukan label), jumlah opsi mengikuti tier (SD=3, SMP=4, SMK/Umum=5).
2. **Essay/Numeric Input** — user mengetik jawaban numerik langsung; sistem mencocokkan dengan `correct_answer_value` (dengan toleransi opsional untuk desimal, jika relevan).

**Aturan randomisasi:**
- Saat sesi challenge dimulai, sistem mengambil `question_count_required` soal secara acak dari bank soal challenge tersebut (`published` saja).
- Urutan soal diacak per sesi.
- Untuk soal multiple choice, urutan opsi jawaban diacak per sesi/per user (bukan disimpan tetap di DB).

**Visibilitas soal (setelah ada organisasi):**
- Soal dengan `organization_id = null` = bank soal publik/global, dibuat admin platform, bisa dipakai siapa saja termasuk guru organisasi (lihat 3.2).
- Soal dengan `organization_id` terisi = privat milik organisasi tersebut, dibuat guru, hanya terlihat & terpakai di dalam organisasi itu.

### 3.5 Gameplay Flow (Challenge Session)

1. User memilih Tier → (Batch/Kategori) → Challenge yang tersedia (unlocked).
2. Sistem mengecek nyawa (`lives_remaining > 0`); jika 0, blokir mulai challenge dan tampilkan waktu reset berikutnya.
3. **Gacha Fetch:** backend mengambil seluruh soal (beserta opsi teracak) untuk sesi ini sekaligus, mengembalikan sebagai satu payload ke client, dan mencatat `challenge_attempt` baru dengan status `in_progress`.
4. Client menjalankan sesi sepenuhnya dari data yang sudah di-fetch (tidak perlu request tambahan per soal).
5. Saat submit, client mengirim seluruh jawaban sekaligus (atau per soal, sesuai keputusan implementasi) ke endpoint submit.
6. Backend menghitung skor, menentukan lulus/tidak berdasarkan `pass_threshold`, mengurangi nyawa jika gagal (aturan pengurangan nyawa perlu diputuskan: per kesalahan atau per kegagalan sesi — direkomendasikan **per kegagalan sesi** untuk konsistensi dengan pola nyawa harian).
7. Jika challenge adalah exam dan lulus → batch berikutnya di-unlock untuk user tersebut.
8. Update streak harian user.

**Alur ujian organisasi (`organization_exams`)** mengikuti pola yang sama persis (start → gacha fetch → snapshot → submit → skor), hanya beda sumber data: soal diambil dari `organization_exam_questions` (bisa fixed set pilihan guru atau random dari pool, lihat `selection_mode` di 4.2), dan attempt dicatat di `organization_exam_attempts`, bukan `challenge_attempts`. Tidak memotong nyawa (`lives`) karena ujian organisasi bukan bagian dari progres kurikulum platform — kecuali diputuskan lain di kemudian hari.

### 3.6 Sistem Nyawa (Lives)

**Aturan:**
- Reset otomatis: +3 nyawa setiap hari (mis. jam 00:00 waktu lokal user atau UTC, perlu diputuskan; direkomendasikan UTC agar konsisten di backend, sementara UI menampilkan estimasi lokal).
- Batas akumulasi maksimum: 15 nyawa per minggu — artinya reset harian tidak menambah nyawa melebihi cap mingguan meski user tidak pernah main.
- User dapat mereset nyawa secara manual dari dashboard mereka sendiri.
  - **Catatan desain terbuka:** perlu diputuskan apakah reset manual ini tanpa batas, dibatasi frekuensi tertentu (misal 1x/hari), atau nantinya jadi salah satu benefit subscription premium. Direkomendasikan: dibatasi (misal 1x per hari) di fase gratis, agar sistem nyawa tetap punya makna kelangkaan.
- Implementasi disarankan dengan job scheduler (cron) harian yang idempotent, plus pencatatan `last_reset_at` per user untuk mencegah reset ganda.

### 3.7 Sistem Streak

- `current_streak`, `longest_streak`, `last_active_date` disimpan per user.
- Streak bertambah jika user menyelesaikan minimal satu challenge (attempt dengan status selesai, lulus atau tidak — perlu diputuskan apakah harus lulus atau cukup mencoba; direkomendasikan: cukup menyelesaikan satu sesi, tidak harus lulus, supaya tetap encourage latihan).
- Streak reset ke 0 jika `last_active_date` lebih dari 1 hari dari hari ini.

### 3.8 Subscription & Entitlement (Skema Future-Proof)

- Skema `plans` dan `user_subscriptions` disiapkan dari awal walau belum ada fitur premium aktif maupun payment gateway.
- Semua user default berada di plan `free` (`is_default = true`), tanpa `expires_at`.
- Entitlement (fitur apa yang terbuka per plan) disimpan sebagai data terpisah (`plan_features`) agar fleksibel ditambah tanpa migrasi skema saat monetisasi mulai digarap.
- Model yang sama nantinya bisa dipakai untuk paket berbayar tingkat **organisasi** (mis. limit jumlah murid/kelas per organisasi) — lihat Open Decisions #8.

### 3.9 Dark Mode & Multi-bahasa

- Dark mode: preferensi disimpan di `users.theme_preference` (`light` / `dark` / `system`), di-sync antar device via akun (bukan cuma localStorage).
- Multi-bahasa: `users.locale_preference`, konten UI di-translate via i18n resource files (Next.js i18n), sementara **konten soal** yang butuh multi-bahasa (jika ada teks soal, bukan cuma angka) disimpan di tabel `question_translations` terpisah dari `questions` agar satu soal bisa punya banyak versi bahasa tanpa duplikasi bank soal.

### 3.10 Multiplayer (Fase 2 — High Level Placeholder)

- Belum didetailkan mekanismenya (real-time via WebSocket vs async battle) — akan dirinci dalam FSD terpisah saat fase 2 dimulai.
- Rekomendasi awal: siapkan tabel `multiplayer_sessions` & `multiplayer_participants` sejak fase 1 jika waktu memungkinkan, agar tidak perlu migrasi besar; namun tidak wajib untuk MVP fase 1.

---

## 4. Skema Database (PostgreSQL)

### 4.1 Autentikasi & User

```sql
users
- id (uuid, pk)
- email (varchar, unique, not null)
- username (varchar, unique, nullable)
- password_hash (varchar, nullable)  -- nullable karena bisa SSO-only
- display_name (varchar)
- role (enum: student, admin, superadmin) default 'student'   -- role PLATFORM, terpisah dari role organisasi
- theme_preference (enum: light, dark, system) default 'system'
- locale_preference (varchar) default 'id'
- current_streak (int) default 0
- longest_streak (int) default 0
- last_active_date (date, nullable)
- created_at, updated_at (timestamptz)

auth_providers
- id (uuid, pk)
- user_id (uuid, fk -> users.id)
- provider (enum: google, password)
- provider_uid (varchar, nullable)  -- google sub id
- created_at (timestamptz)
- unique (provider, provider_uid)
```

### 4.2 Organisasi & Grouping

```sql
organizations
- id (uuid, pk)
- name (varchar)
- slug (varchar, unique)
- type (enum: school, other) default 'school'
- owner_user_id (uuid, fk -> users.id)
- status (enum: active, suspended) default 'active'
- created_at, updated_at

organization_members
- id (uuid, pk)
- organization_id (uuid, fk -> organizations.id)
- user_id (uuid, fk -> users.id, nullable)      -- null selama invite belum di-accept
- invited_email (varchar, nullable)             -- dipakai sebelum user_id terisi
- role (enum: owner, admin, teacher, student)
- status (enum: invited, active, removed) default 'invited'
- invited_by (uuid, fk -> users.id, nullable)
- joined_at (timestamptz, nullable)
- created_at, updated_at
- unique (organization_id, user_id) where user_id is not null
- unique (organization_id, invited_email) where status = 'invited'

organization_invites
- id (uuid, pk)
- organization_id (uuid, fk -> organizations.id)
- organization_member_id (uuid, fk -> organization_members.id)
- group_id (uuid, fk -> groups.id, nullable)     -- kelas tujuan, jika undangan untuk murid
- token (varchar, unique)
- expires_at (timestamptz)
- accepted_at (timestamptz, nullable)
- created_at (timestamptz)

groups                                           -- grouping/kelas, mendukung hierarki 2 level
- id (uuid, pk)
- organization_id (uuid, fk -> organizations.id)
- parent_group_id (uuid, fk -> groups.id, nullable)  -- mis. "Kelas 10" sbg parent dari "Kelas 10 A"
- name (varchar)
- order_index (int)
- created_at, updated_at

group_members
- id (uuid, pk)
- group_id (uuid, fk -> groups.id)
- organization_member_id (uuid, fk -> organization_members.id)
- role_in_group (enum: teacher, student)
- created_at (timestamptz)
- unique (group_id, organization_member_id)
```

### 4.3 Kurikulum

```sql
tiers
- id (uuid, pk)
- code (varchar, unique)        -- 'sd', 'smp', 'smk', 'kuliah', 'umum'
- name (varchar)
- uses_batch (boolean)          -- false khusus untuk 'umum'
- order_index (int)

batches
- id (uuid, pk)
- tier_id (uuid, fk -> tiers.id)
- name (varchar)                -- 'Batch 1 - Kelas 1'
- order_index (int)
- created_at, updated_at

categories                      -- khusus tier 'umum'
- id (uuid, pk)
- tier_id (uuid, fk -> tiers.id)
- name (varchar)                -- 'Crypto Arithmetic'
- order_index (int)

challenges
- id (uuid, pk)
- batch_id (uuid, fk -> batches.id, nullable)
- category_id (uuid, fk -> categories.id, nullable)
  -- constraint: tepat salah satu dari batch_id / category_id yang terisi
- name (varchar)
- order_index (int)
- is_exam (boolean) default false
- question_count_required (int)     -- misal 20
- pass_threshold_percent (int)      -- misal 80
- option_count (int)                -- 3 / 4 / 5, default mengikuti tier
- created_at, updated_at
```

### 4.4 Bank Soal

```sql
questions
- id (uuid, pk)
- challenge_id (uuid, fk -> challenges.id, nullable)     -- nullable: soal organisasi belum tentu menempel ke challenge platform
- organization_id (uuid, fk -> organizations.id, nullable)  -- null = bank soal publik/global; terisi = privat milik organisasi tsb
- question_type (enum: multiple_choice, essay_numeric)
- prompt (text)                     -- teks/angka soal, mis. "2 x 2"
- correct_answer_value (numeric)
- status (enum: draft, published, archived) default 'draft'
- created_by (uuid, fk -> users.id)
- created_at, updated_at

question_options                    -- hanya untuk question_type = multiple_choice
- id (uuid, pk)
- question_id (uuid, fk -> questions.id)
- option_value (numeric)
- is_correct (boolean)

question_translations               -- opsional, jika prompt butuh multi-bahasa
- id (uuid, pk)
- question_id (uuid, fk -> questions.id)
- locale (varchar)
- prompt_text (text)
```

> Catatan: `challenge_id` di `questions` dibuat nullable karena soal buatan guru organisasi tidak selalu terikat ke `challenge` platform — soal tersebut dipakai lewat `organization_exam_questions` (lihat 4.5), bukan lewat hierarki tier/batch/challenge.

### 4.5 Progres & Attempt

```sql
user_progress
- id (uuid, pk)
- user_id (uuid, fk -> users.id)
- tier_id (uuid, fk -> tiers.id)
- current_batch_id (uuid, fk -> batches.id, nullable)
- unlocked_challenge_ids (uuid[])       -- atau tabel relasi terpisah, tergantung preferensi query
- updated_at

challenge_attempts
- id (uuid, pk)
- user_id (uuid, fk -> users.id)
- challenge_id (uuid, fk -> challenges.id)
- status (enum: in_progress, passed, failed)
- score_percent (int, nullable)
- questions_snapshot (jsonb)        -- snapshot soal+opsi teracak yang di-fetch di awal sesi (gacha)
- answers_submitted (jsonb, nullable)
- started_at, completed_at (timestamptz)

lives
- user_id (uuid, pk, fk -> users.id)
- lives_remaining (int) default 3
- weekly_lives_used_or_cap_tracker (int)  -- untuk hitung cap 15/minggu
- last_daily_reset_at (timestamptz)
- last_manual_reset_at (timestamptz, nullable)
- week_start_at (date)               -- penanda periode mingguan untuk cap

organization_exams
- id (uuid, pk)
- organization_id (uuid, fk -> organizations.id)
- group_id (uuid, fk -> groups.id, nullable)   -- null = berlaku untuk seluruh organisasi
- created_by (uuid, fk -> users.id)            -- guru pembuat
- name (varchar)
- description (text, nullable)
- selection_mode (enum: fixed_set, random_from_pool) default 'fixed_set'
- question_count_required (int, nullable)      -- dipakai jika selection_mode = random_from_pool
- pass_threshold_percent (int, nullable)
- option_count (int) default 4
- status (enum: draft, published, archived) default 'draft'
- opens_at (timestamptz, nullable)
- closes_at (timestamptz, nullable)
- created_at, updated_at

organization_exam_questions
- id (uuid, pk)
- organization_exam_id (uuid, fk -> organization_exams.id)
- question_id (uuid, fk -> questions.id)       -- bisa soal publik (organization_id null) atau soal privat organisasi ini
- order_index (int, nullable)
- unique (organization_exam_id, question_id)

organization_exam_attempts
- id (uuid, pk)
- organization_exam_id (uuid, fk -> organization_exams.id)
- user_id (uuid, fk -> users.id)
- status (enum: in_progress, passed, failed)
- score_percent (int, nullable)
- questions_snapshot (jsonb)
- answers_submitted (jsonb, nullable)
- started_at, completed_at (timestamptz)
```

> Catatan: `questions_snapshot` (jsonb) penting untuk pola "gacha fetch" — begitu sesi dimulai, soal+opsi yang sudah diacak disimpan sebagai snapshot immutable di attempt tersebut, sehingga scoring saat submit selalu konsisten dengan apa yang dilihat user, terlepas dari perubahan bank soal setelahnya. Berlaku sama untuk `challenge_attempts` maupun `organization_exam_attempts`.

### 4.6 Subscription (Future-Proof, Belum Aktif)

```sql
plans
- id (uuid, pk)
- code (varchar, unique)        -- 'free', 'premium' (contoh)
- name (varchar)
- is_default (boolean) default false
- price_amount (numeric, nullable)
- price_currency (varchar, nullable)
- billing_period (enum: monthly, yearly, none, nullable)

plan_features
- id (uuid, pk)
- plan_id (uuid, fk -> plans.id)
- feature_key (varchar)          -- 'extra_lives', 'multiplayer_ranked', dst
- feature_value (jsonb, nullable)

user_subscriptions
- id (uuid, pk)
- user_id (uuid, fk -> users.id)
- plan_id (uuid, fk -> plans.id)
- status (enum: active, canceled, expired) default 'active'
- started_at (timestamptz)
- expires_at (timestamptz, nullable)
```

---

## 5. Daftar Endpoint API (High-Level, REST)

### Auth
- `POST /auth/register` — register email/username + password
- `POST /auth/login` — login password
- `GET /auth/google` / `GET /auth/google/callback` — OAuth flow
- `POST /auth/set-password` — set password untuk akun SSO
- `POST /auth/refresh`
- `POST /auth/logout`

### Organization
- `POST /organizations/register` — daftar organisasi baru + akun owner sekaligus
- `GET /organizations/:id`
- `PATCH /organizations/:id`
- `GET /organizations/:id/members`
- `POST /organizations/:id/invites` — undang guru/murid (email, role, group_id opsional)
- `PATCH /organization-members/:id` — ubah role/status, keluarkan member
- `GET /invites/:token` — validasi undangan (dipakai halaman accept-invite sebelum redirect ke Google)
- `POST /invites/:token/accept` — accept undangan setelah Google OAuth berhasil
- `POST /organizations/:id/groups` — buat grouping/kelas (support `parent_group_id`)
- `GET /organizations/:id/groups`
- `PATCH/DELETE /groups/:id`

### Organization Exam
- `POST /organizations/:id/exams` — buat ujian
- `GET /organizations/:id/exams` — list ujian (filter by group)
- `PATCH/DELETE /exams/:id`
- `POST /exams/:id/questions` — tambah soal ke ujian (dari bank publik via `question_id`, atau bikin soal privat baru)
- `GET /question-bank?scope=global` — guru browse bank soal publik untuk dipakai di ujiannya
- `POST /exams/:id/start` — gacha fetch, buat `organization_exam_attempt`
- `POST /organization-exam-attempts/:id/submit`

### Curriculum (Public/Student)
- `GET /tiers`
- `GET /tiers/:tierId/batches` (atau `/categories` untuk tier umum)
- `GET /batches/:batchId/challenges`
- `GET /challenges/:challengeId` — detail challenge (belum termasuk soal)

### Gameplay
- `POST /challenges/:challengeId/start` — gacha fetch, membuat `challenge_attempt`, cek & potong nyawa jika perlu
- `POST /attempts/:attemptId/submit` — submit jawaban, hitung skor, update progres/streak/nyawa

### User
- `GET /me`
- `GET /me/progress`
- `GET /me/lives`
- `GET /me/organizations` — daftar organisasi tempat user jadi member (owner/admin/teacher/student)
- `POST /me/lives/reset` — reset manual (dengan aturan pembatasan sesuai 3.6)
- `PATCH /me/preferences` — theme & locale

### Admin (Platform)
- `POST/PUT/DELETE /admin/tiers`
- `POST/PUT/DELETE /admin/batches`
- `POST/PUT/DELETE /admin/categories`
- `POST/PUT/DELETE /admin/challenges`
- `POST/PUT/DELETE /admin/questions`
- `POST/PUT/DELETE /admin/questions/:id/options`
- `GET /admin/organizations` — list & kelola organisasi terdaftar
- `GET /admin/stats` — statistik agregat (jumlah soal per tier, completion rate, dll)

---

## 6. Hal yang Perlu Diputuskan Lebih Lanjut (Open Decisions)

1. Aturan pasti pengurangan nyawa: per kegagalan sesi vs per jawaban salah.
2. Batas frekuensi reset nyawa manual (bebas / dibatasi per hari).
3. Timezone acuan untuk reset harian nyawa & streak (rekomendasi: UTC di backend, tampilan lokal di frontend).
4. Apakah streak membutuhkan "lulus" challenge atau cukup "menyelesaikan" sesi.
5. Apakah soal essay/numeric butuh toleransi (misal untuk jawaban desimal/pecahan) atau harus exact match.
6. Mekanisme multiplayer fase 2 (real-time vs async) — akan menentukan kebutuhan infrastruktur tambahan (WebSocket server, dsb).
7. Apakah satu murid bisa jadi anggota lebih dari satu organisasi/kelas sekaligus (mis. pindah sekolah, ikut bimbel tambahan)? Rekomendasi: boleh, tidak dibatasi di level skema.
8. Apakah organisasi punya limit jumlah anggota/kelas berdasarkan plan berbayar (tie-in ke skema subscription di 3.8)?
9. Apakah soal privat buatan guru bisa "dipromosikan" jadi bagian bank soal publik oleh admin platform, dan lewat mekanisme apa (review manual, dsb)?
10. Provider & template untuk pengiriman email undangan organisasi (SendGrid/SES/lainnya) — belum ditentukan.
11. Apakah perlu role `admin` organisasi yang terpisah wewenangnya dari `owner` secara lebih granular (mis. staf tata usaha vs kepala sekolah), atau cukup dua level (`owner`/`admin`) seperti yang didefinisikan sekarang.
12. Apakah `organization_exam_attempts` perlu ikut mempengaruhi `current_streak` platform, atau streak murni dari `challenge_attempts` saja.
