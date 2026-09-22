# syntax=docker/dockerfile:1

# Build semua binary cmd/* dalam satu stage biar image akhir gak bawa toolchain
# Go -- cuma binary + migrations/ yang kepake pas runtime (lihat cmd/migrate,
# yang baca migrations/*.sql relatif ke working directory).
FROM golang:1.26-alpine AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN mkdir -p /out/bin && \
    for cmd in api migrate seed seed-curriculum seed-essay seed-puzzle; do \
      CGO_ENABLED=0 GOOS=linux go build -o /out/bin/"$cmd" ./cmd/"$cmd"; \
    done

FROM alpine:3.20
RUN apk add --no-cache ca-certificates wget
WORKDIR /app

COPY --from=builder /out/bin/ ./bin/
COPY migrations ./migrations

ENV PORT=9801
EXPOSE 9801

HEALTHCHECK --interval=30s --timeout=3s --retries=3 \
    CMD wget -qO- http://localhost:9801/health || exit 1

# Binary lain (migrate/seed*) dijalanin one-off lewat:
#   docker compose run --rm api /app/bin/migrate
ENTRYPOINT ["/app/bin/api"]
