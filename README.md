# MathQuest — Monorepo

Lihat [BRD.md](BRD.md), [PRD.md](PRD.md), dan [FSD.md](FSD.md) untuk spesifikasi lengkap.

## Struktur

```
lesson/
├─ apps/
│  ├─ web/          # Next.js — Lesson App (userApp), diakses end-user
│  ├─ dashboard/     # Next.js — dashboard organisasi + dashboard admin platform
│  └─ mobile/        # Expo (React Native) — Lesson App versi mobile, mirror apps/web
├─ services/
│  └─ api/           # Go — backend API
└─ docker-compose.yml # Postgres lokal
```

## Port (dev lokal)

| Service | Port |
|---|---|
| Postgres (DB) | 9800 |
| Go API | 9801 |
| Dashboard (Next.js) | 9802 |
| userApp (Next.js) | 9803 |
| Mobile (Expo web, dev/testing) | 9804 |

## Menjalankan

```bash
# 1. Database
docker compose up -d

# 2. Backend API
cp services/api/.env.example services/api/.env
cd services/api && go run ./cmd/api

# 3. userApp
cd apps/web && npm run dev

# 4. Dashboard
cd apps/dashboard && npm run dev

# 5. Mobile (Expo)
cp apps/mobile/.env.example apps/mobile/.env
cd apps/mobile && npx expo start --web --port 9804
```
