# Panduan Pengembangan Lokal

Panduan ini menjelaskan workflow pengembangan harian, pengelolaan database, troubleshooting, dan setup IDE untuk project **golang-api-hexagonal**.

---

## Daftar Isi

- [Quick Start (TL;DR)](#quick-start-tldr)
- [Prerequisites](#prerequisites)
- [Daily Workflow Commands](#daily-workflow-commands)
  - [Menjalankan Aplikasi](#menjalankan-aplikasi)
  - [Menjalankan Tests](#menjalankan-tests)
  - [Linting](#linting)
  - [CI Pipeline Lokal](#ci-pipeline-lokal)
- [Hot Reload](#hot-reload)
- [Database Management](#database-management)
  - [Menjalankan PostgreSQL](#menjalankan-postgresql)
  - [Database Migrations](#database-migrations)
  - [Environment Variables](#environment-variables)
- [Troubleshooting](#troubleshooting)
- [IDE Setup](#ide-setup)
  - [VS Code](#vs-code)
  - [GoLand](#goland)
- [Referensi Lengkap Make Commands](#referensi-lengkap-make-commands)

---

## Quick Start (TL;DR)

```bash
# 1. Setup tools dan dependencies
make setup

# 2. Jalankan PostgreSQL
make infra-up

# 3. Set DATABASE_URL
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/golang_api_hexagonal?sslmode=disable"

# 4. Jalankan migrations
make migrate-up

# 5. Jalankan aplikasi
make run
```

> [!TIP]
> Server akan berjalan di `http://localhost:8080`. Cek health endpoint: `curl http://localhost:8080/health`

---

## Prerequisites

Sebelum memulai, pastikan tools berikut sudah terinstall:

| Tool | Versi Minimum | Cek Versi |
|------|---------------|-----------|
| **Go** | 1.24+ | `go version` |
| **Docker** | 20.10+ | `docker --version` |
| **Docker Compose** | 2.0+ | `docker compose version` |

> [!NOTE]
> Jalankan `make setup` untuk menginstall tools development seperti `golangci-lint` dan `goose` secara otomatis.

---

## Daily Workflow Commands

### Menjalankan Aplikasi

```bash
# Jalankan aplikasi (development mode)
make run
```

**Output yang diharapkan:**
```
{"time":"2025-12-23T00:00:00Z","level":"INFO","msg":"starting server","addr":":8080"}
```

Aplikasi berjalan di `http://localhost:8080`.

**Endpoints yang tersedia:**
- `GET /health` - Health check
- `GET /ready` - Readiness check
- `GET /metrics` - Prometheus metrics

---

### Menjalankan Tests

```bash
# Jalankan semua tests dengan race detection
make test
```

Perintah ini menjalankan:
- Semua unit tests
- Race condition detection (`-race` flag)
- Coverage profiling

**Cek coverage threshold (80% untuk domain+app):**

```bash
make coverage
```

**Output yang diharapkan:**
```
📊 Running tests with coverage (domain+app)...
📈 Coverage report:
total:    (statements)    85.5%
✅ Coverage 85.5% meets 80% threshold
```

---

### Linting

```bash
# Jalankan golangci-lint
make lint
```

Linter dikonfigurasi di `.golangci.yml` dengan rules:
- Layer boundary enforcement (depguard)
- Code style checks
- Security scanning

> [!WARNING]
> Lint errors akan menyebabkan CI gagal. Selalu jalankan `make lint` sebelum commit.

---

### CI Pipeline Lokal

```bash
# Jalankan full CI pipeline secara lokal
make ci
```

Pipeline mencakup:
1. `check-mod-tidy` - Verifikasi go.mod tidy
2. `check-fmt` - Verifikasi formatting (gofmt)
3. `lint` - Jalankan golangci-lint
4. `test` - Jalankan semua tests

> [!TIP]
> Jalankan `make ci` sebelum push untuk memastikan CI tidak gagal.

---

## Hot Reload

Go tidak memiliki built-in hot reload. Berikut opsi yang tersedia:

### Opsi 1: Manual Restart (Recommended)

Workflow paling sederhana dan reliable:

```bash
# Terminal 1: Stop dengan Ctrl+C, lalu restart
make run
```

> [!TIP]
> Gunakan shortcut shell untuk restart cepat: tekan `Ctrl+C`, lalu `↑` (arrow up) + `Enter`.

### Opsi 2: Menggunakan Air (Auto-reload)

[Air](https://github.com/cosmtrek/air) adalah live reload tool untuk Go.

**Install Air:**
```bash
go install github.com/cosmtrek/air@latest
```

**Jalankan dengan Air:**
```bash
air
```

> [!CAUTION]
> Air tidak termasuk dalam setup default project. Jika menggunakan Air, buat file `.air.toml` untuk konfigurasi.

### Opsi 3: Menggunakan Entr (File watcher)

```bash
# Install entr (macOS)
brew install entr

# Watch dan restart saat file berubah
find . -name '*.go' | entr -r make run
```

---

## Database Management

### Menjalankan PostgreSQL

```bash
# Start PostgreSQL container
make infra-up
```

**Output:**
```
🐘 Starting PostgreSQL...
⏳ Waiting for PostgreSQL to be healthy (timeout: 60s)...
✅ Infrastructure is ready!

PostgreSQL connection:
  Host: localhost:5432
  User: postgres
  Pass: postgres
  DB:   golang_api_hexagonal
```

**Stop PostgreSQL (data tetap tersimpan):**
```bash
make infra-down
```

**Reset PostgreSQL (HAPUS SEMUA DATA):**
```bash
make infra-reset
```

> [!CAUTION]
> `make infra-reset` akan menghapus semua data database! Gunakan dengan hati-hati.

**Cek status infrastructure:**
```bash
make infra-status
```

**Lihat logs PostgreSQL:**
```bash
make infra-logs
```

---

### Database Migrations

**Set DATABASE_URL terlebih dahulu:**

```bash
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/golang_api_hexagonal?sslmode=disable"
```

| Command | Deskripsi |
|---------|-----------|
| `make migrate-up` | Jalankan semua pending migrations |
| `make migrate-down` | Rollback migration terakhir |
| `make migrate-status` | Lihat status migrations |
| `make migrate-create name=description` | Buat migration file baru |
| `make migrate-validate` | Validasi syntax migration files |

**Contoh membuat migration baru:**

```bash
make migrate-create name=add_user_roles
```

**Output:**
```
Created: migrations/20251223123456_add_user_roles.sql
```

> [!IMPORTANT]
> Migration files harus memiliki section `-- +goose Up` dan `-- +goose Down`.

---

### Environment Variables

**Required environment variables:**

```bash
# Database connection
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/golang_api_hexagonal?sslmode=disable"
```

**Cara cepat dari .env.example:**

```bash
# Source dari .env.example
export $(grep DATABASE_URL .env.example | xargs)
```

> [!TIP]
> Buat file `.env` lokal (sudah di-gitignore) untuk menyimpan environment variables.

---

## Troubleshooting

### Port 8080 Already in Use

**Gejala:**
```
listen tcp :8080: bind: address already in use
```

**Solusi:**

```bash
# Cari proses yang menggunakan port 8080
lsof -i :8080

# Kill proses tersebut
kill -9 <PID>
```

Atau jalankan di port lain:
```bash
PORT=8081 make run
```

---

### Database Connection Refused

**Gejala:**
```
dial tcp 127.0.0.1:5432: connect: connection refused
```

**Solusi:**

1. Pastikan PostgreSQL berjalan:
   ```bash
   make infra-status
   ```

2. Jika tidak berjalan, start ulang:
   ```bash
   make infra-up
   ```

3. Tunggu sampai healthy (biasanya ~10 detik)

---

### Migration Errors

**Gejala:**
```
❌ DATABASE_URL is not set.
```

**Solusi:**
```bash
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/golang_api_hexagonal?sslmode=disable"
```

**Gejala:**
```
❌ goose not found. Run 'make setup' first.
```

**Solusi:**
```bash
make setup
```

---

### Docker/PostgreSQL Container Not Starting

**Gejala:**
```
❌ PostgreSQL reported unhealthy
```

**Solusi:**

1. Cek logs:
   ```bash
   make infra-logs
   ```

2. Reset dan coba lagi:
   ```bash
   make infra-reset INFRA_CONFIRM=y
   make infra-up
   ```

3. Pastikan Docker daemon berjalan:
   ```bash
   docker info
   ```

---

### Golangci-lint Version Mismatch

**Gejala:**
```
Error: cannot load package: package ... module requires Go 1.24
```

**Solusi:**
```bash
# Update golangci-lint ke versi terbaru
go install github.com/golangci-lint/golangci-lint/cmd/golangci-lint@latest

# Verifikasi versi
golangci-lint --version
```

---

### Test Failures with Race Detection

**Gejala:**
```
WARNING: DATA RACE
```

**Solusi:**

1. Race condition terdeteksi di code - FIX REQUIRED
2. Cek output untuk mengetahui file dan line yang bermasalah
3. Gunakan mutex atau channel untuk synchronization

> [!IMPORTANT]
> Jangan disable race detection! Race conditions adalah bugs serius.

---

### Environment Variable Not Set Errors

**Gejala:**
```
⚠️ Required configuration not set
```

**Solusi:**

1. Cek `.env.example` untuk required variables
2. Set semua required variables:
   ```bash
   export DATABASE_URL="..."
   # ... other required vars
   ```

---

## IDE Setup

### VS Code

**1. Install Go Extension:**

- Buka Extensions (Ctrl+Shift+X)
- Cari "Go" oleh Go Team at Google
- Install

**2. Recommended Settings:**

Buat file `.vscode/settings.json`:

```json
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.lintFlags": [
    "--fast"
  ],
  "go.formatTool": "gofmt",
  "go.testFlags": ["-race"],
  "go.coverOnSave": true,
  "go.coverageDecorator": {
    "type": "highlight"
  },
  "editor.formatOnSave": true,
  "[go]": {
    "editor.codeActionsOnSave": {
      "source.organizeImports": "explicit"
    }
  }
}
```

**3. Debug Configuration:**

Buat file `.vscode/launch.json`:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Launch API",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${workspaceFolder}/cmd/api",
      "env": {
        "DATABASE_URL": "postgres://postgres:postgres@localhost:5432/golang_api_hexagonal?sslmode=disable"
      }
    },
    {
      "name": "Debug Test",
      "type": "go",
      "request": "launch",
      "mode": "test",
      "program": "${fileDirname}",
      "args": ["-test.run", "${input:testName}"]
    }
  ],
  "inputs": [
    {
      "id": "testName",
      "type": "promptString",
      "description": "Test function name"
    }
  ]
}
```

**4. Golangci-lint Integration:**

Golangci-lint akan otomatis berjalan jika:
- Setting `go.lintTool` = "golangci-lint"
- File `.golangci.yml` ada di root project

---

### GoLand

**1. Open Project:**

- File → Open → Pilih folder project
- GoLand akan otomatis detect Go project

**2. Golangci-lint Integration:**

- File → Settings → Tools → File Watchers
- Klik "+" → Custom
- Konfigurasi:
  - Name: `golangci-lint`
  - Program: `golangci-lint`
  - Arguments: `run --fast $FileDir$`
  - Output paths: `$ProjectFileDir$`
  - Working directory: `$ProjectFileDir$`

**3. Database Tool:**

- View → Tool Windows → Database
- Klik "+" → Data Source → PostgreSQL
- Konfigurasi:
  - Host: `localhost`
  - Port: `5432`
  - User: `postgres`
  - Password: `postgres`
  - Database: `golang_api_hexagonal`

**4. Run/Debug Configuration:**

- Run → Edit Configurations
- Klik "+" → Go Build
- Konfigurasi:
  - Name: `API Server`
  - Run kind: `Package`
  - Package path: `github.com/iruldev/golang-api-hexagonal/cmd/api`
  - Environment: `DATABASE_URL=postgres://postgres:postgres@localhost:5432/golang_api_hexagonal?sslmode=disable`

**5. Test dengan Race Detection:**

- Preferences → Go → Go Test
- Additional Test Options: `-race`

---

## Referensi Lengkap Make Commands

| Command | Deskripsi |
|---------|-----------|
| **Development** | |
| `make setup` | Install development tools dan dependencies |
| `make build` | Build application binary |
| `make run` | Jalankan aplikasi dengan `go run ./cmd/api` |
| `make test` | Jalankan semua tests dengan race detection |
| `make coverage` | Cek coverage threshold 80% (domain+app) |
| `make lint` | Jalankan golangci-lint |
| `make clean` | Hapus build artifacts |
| **CI Pipeline** | |
| `make ci` | Jalankan full CI pipeline lokal |
| `make check-mod-tidy` | Verifikasi go.mod tidy |
| `make check-fmt` | Verifikasi code formatting |
| **Infrastructure** | |
| `make infra-up` | Start PostgreSQL container |
| `make infra-down` | Stop PostgreSQL (preserve data) |
| `make infra-reset` | Reset all infrastructure (DELETE data) |
| `make infra-logs` | View PostgreSQL logs |
| `make infra-status` | Show infrastructure status |
| **Migrations** | |
| `make migrate-up` | Jalankan pending migrations |
| `make migrate-down` | Rollback last migration |
| `make migrate-status` | Show migration status |
| `make migrate-create name=x` | Buat migration file baru |
| `make migrate-validate` | Validasi migration syntax |

---

## Struktur Project

```
.
├── cmd/api/main.go              # Entry point
├── internal/
│   ├── domain/                  # Business entities dan interfaces
│   ├── app/                     # Use cases
│   ├── transport/http/          # HTTP handlers dan middleware
│   └── infra/                   # External implementations
├── migrations/                  # Goose SQL migration files
├── docs/                        # Documentation
├── Makefile                     # Development commands
├── docker-compose.yml           # Infrastructure
├── .golangci.yml                # Linter configuration
└── .env.example                 # Environment template
```

---

**Last Updated:** 2025-12-23
