# 🚂 RelGo Framework: Opinionated Go Backend Framework

**RelGo** (rel) adalah **opinionated backend framework** untuk Go (Golang) yang dibangun di atas **Fiber v3**, **GORM**, dan **Viper**. Filosofinya seperti **Rel Kereta Api**: developer tinggal berjalan di atas jalur rel yang sudah disiapkan, sehingga dipastikan sampai ke tujuan (*production-ready API*) dengan cepat tanpa resiko anjlok (*architectural erosion*).

> [!IMPORTANT]
> **📜 Kontrak Arsitektur Framework:**
> Dengan **RelGo**, untuk membuat sebuah modul baru, developer idealnya **hanya perlu menyentuh 3 area**:
> 1. 📦 **Entity / Model** *(Definisi domain data & kolom)*
> 2. 🗄️ **Repository Query** *(Query spesifik database)*
> 3. 🧠 **Business Rule** *(Logika bisnis di Service layer)*
>
> Semua sisanya (*HTTP routing, format response, DTO request/response, mapper, validasi input dasar, logging, dependency wiring, dan dokumentasi Swagger*) ditangani secara otomatis oleh **Foundation** dan **rel CLI Generator**.

---

## 📊 Compatibility Matrix

| Komponen | Versi Teruji & Didukung | Keterangan |
| :--- | :--- | :--- |
| **Go Compiler** | `Go 1.25.0+` | Minimal versi Go yang dibutuhkan |
| **Web Engine** | `Fiber v3 (v3.4.0)` | High-performance Fasthttp web engine |
| **ORM / Database** | `GORM v1.31+` / `MySQL 8.0` | Driver MySQL / MariaDB |
| **Cache & Storage** | `Redis v7.0` | Storage rate limiter & caching |
| **Config Loader** | `Viper v1.21` | Fail-fast JSON/Env configuration |
| **API Docs** | `Swagger / Swaggo v1.16` | OpenAPI 2.0 automatic documentation |

---

## 🗺️ Public Roadmap

```text
v1.0 (Core Frozen & Production Ready) - SELESAI ✅
 ├── Foundation Architecture (Response, Exception, Mapper, Validator, Context)
 ├── BaseEntity & Query Filter (Paginasi, Search & Sort)
 ├── Health Checks (/health, /ready, /live) & Graceful Shutdown
 ├── Docker Compose Setup (MySQL 8 & Redis 7)
 ├── Architecture Linter (go run ./cmd/lint)
 ├── Starter Kit Doctor (go run ./cmd/doctor)
 └── Unified CLI Tooling (rel) & Golden Snapshot Testing

v1.1 (CLI Distribution & DX Polish) - NEXT 🚀
 ├── Global Binary Installer: `go install github.com/mkhsnw/rel@latest`
 ├── Pretty Colorized Terminal Scaffolding (`rel new <project-name>`)
 └── Multi-Domain Example Repositories (`examples/library-api`)

v1.2 (Interactive Engine & Manifest Plugins)
 ├── Interactive Terminal Wizard Enhancements
 └── Custom Generator Template Overrides
```

---

## 📖 Daftar Isi

1. [Fitur Utama & Keunggulan](#-fitur-utama--keunggulan)
2. [Instalasi & Penggunaan CLI (`rel`)](#-instalasi--penggunaan-cli-rel)
3. [Filosofi Arsitektur (The "Why")](#-filosofi-arsitektur-the-why)
4. [Fitur Produksi & Keamanan](#-fitur-produksi--keamanan)
5. [Developer Experience & Quality Guard](#-developer-experience--quality-guard)
   - [🔥 Architecture Linter (`rel lint`)](#-architecture-linter-rel-lint)
   - [🩺 Starter Kit Doctor (`rel doctor`)](#-starter-kit-doctor-rel-doctor)
   - [🧙‍♂️ Interactive Module Generator (`rel gen`)](#-interactive-module-generator-rel-gen)
6. [Struktur Proyek](#-struktur-proyek)
7. [Panduan Generator & Format Manifest](#-panduan-generator--format-manifest)
8. [Format Response API & Error Handling](#-format-response-api--error-handling)
9. [Semua Perintah Task Runner](#-semua-perintah-task-runner)

---

## ✨ Fitur Utama & Keunggulan

- 🏛️ **Clean Architecture Enforcement**: Batasan layer yang ketat (Controller ➔ Service ➔ Repository) yang dijaga secara otomatis oleh **Architecture Linter** berbasis AST Parser (`rel lint`).
- ⚡ **High Performance Stack**: Berbasis [Fiber v3](https://gofiber.io/) (Fasthttp), [GORM](https://gorm.io/), [Logrus](https://github.com/sirupsen/logrus), dan [Viper](https://github.com/spf13/viper).
- 🛡️ **Fail-Fast Configuration**: Validasi otomatis saat aplikasi pertama kali start. Aplikasi langsung *exit* secara aman jika ada konfigurasi environment wajib yang hilang.
- 🩺 **Built-in Health Checks**: Ready-to-use `/health`, `/ready` (ping MySQL & Redis), dan `/live` untuk Kubernetes / Docker Swarm probe.
- 🔄 **Context-Propagated Transactions**: Transaksi database tanpa perlu mengoper struct `tx` secara manual ke signature repository.
- 🔍 **Query Filter Standard**: Filter pencarian (`?search=...`), sorting dinamis (`?sort_by=created_at&sort_dir=desc`), dan paginasi standar terintegrasi di level generic repository.

---

## 🧰 Instalasi & Penggunaan CLI (`rel`)

Kompilasi CLI `rel` secara lokal atau jalankan via Task runner:

```bash
# Build binary CLI rel
go build -o bin/rel.exe ./cmd/rel

# Menjalankan diagnostik kesehatan proyek
./bin/rel.exe doctor   # atau: task doctor

# Menjalankan Architecture Linter
./bin/rel.exe lint     # atau: task lint-arch

# Membuat project baru
./bin/rel.exe new github.com/username/my-api

# Generate modul baru
./bin/rel.exe gen Product
```

---

## 🧠 Filosofi Arsitektur (The "Why")

### 1. Kenapa ada `internal/foundation`?
`foundation` adalah tulang punggung framework. Alih-alih membuat logic paginasi, format error, atau struktur response secara berulang di setiap modul, semuanya ditarik ke tengah. `foundation` memastikan **seluruh API berbicara dengan bahasa yang sama**.

### 2. Kenapa DTO dipisah dari Entity?
Entity adalah representasi skema database, sedangkan DTO adalah kontrak payload transport HTTP. Memisahkan keduanya menjamin field sensitif (seperti `password_hash` atau data audit internal) tidak akan pernah bocor ke response API secara tidak sengaja.

### 3. Thin Controller & Isolated Service
Controller hanya bertugas menerima HTTP request (Fiber Ctx), memanggil Service, dan mengembalikan HTTP Response via `foundation/response`. Service murni berisi logika bisnis yang **bebas dari dependensi web framework Fiber**.

---

## 🚀 Fitur Produksi & Keamanan

### 1. Config Fail-Fast Validation
Aplikasi divalidasi secara otomatis pada startup (`main.go`):
```go
// internal/config/config.go
type JwtConfig struct {
    Secret string `mapstructure:"secret" validate:"required,min=16"`
}
```
Jika `JWT.Secret` kosong atau kurang dari 16 karakter, aplikasi akan mencetak pesan error terperinci dan melakukan `os.Exit(1)`.

### 2. Health & Readiness Probes
- **`GET /health`** (Liveness Probe): Mengembalikan `200 OK`.
- **`GET /ready`** (Readiness Probe): Ping aktif ke **MySQL** dan **Redis**. Mengembalikan `200 OK` (sehat) atau `503 Service Unavailable` (terputus).
- **`GET /live`** (Diagnostic Probe): Status terperinci sistem.

### 3. Graceful Shutdown
Ketika aplikasi menerima sinyal `SIGINT` (Ctrl+C) atau `SIGTERM`, server menyelesaikan request yang sedang berjalan dan menutup koneksi MySQL (`gorm.DB`) dan Redis Storage secara rapi.

### 4. BaseEntity & Core Audit Timestamps
Semua model domain meng-embed `database.BaseEntity`:
```go
package product

import "github.com/mkhsnw/rel/internal/foundation/database"

type Product struct {
    database.BaseEntity // Memiliki ID (UUID v7), CreatedAt, UpdatedAt, DeletedAt (gorm.DeletedAt)
    Name  string  `gorm:"column:name;type:varchar(255);not null"`
    Price float64 `gorm:"column:price;type:double;not null"`
}
```

### 5. Query Object Pattern (Paginasi, Search & Sort)
```go
filter := database.QueryFilter{
    Page:    1,
    Size:    10,
    Search:  "laptop",
    SortBy:  "price",
    SortDir: "asc",
}

products, total, err := r.List(ctx, filter, []string{"name", "sku"})
```

### 6. Context-Based Transaction Manager
```go
err := s.Tx.Run(ctx, func(txCtx context.Context) error {
    if err := s.OrderRepo.Create(txCtx, order); err != nil {
        return err
    }
    if err := s.InventoryRepo.DeductStock(txCtx, productID, qty); err != nil {
        return err
    }
    return nil
})
```

---

## 🛠️ Developer Experience & Quality Guard

### 🔥 Architecture Linter (`rel lint` / `task lint-arch`)
AST Static Analyzer yang mengecek 5 aturan arsitektur secara otomatis:

```bash
task lint-arch
```

### 🩺 Starter Kit Doctor (`rel doctor` / `task doctor`)
Diagnostik lingkungan dev:

```bash
task doctor
```

---

## 📁 Struktur Proyek

```text
├── cmd/
│   ├── main.go             # Entrypoint aplikasi HTTP server
│   ├── rel/                # Unified RelGo CLI Executable
│   ├── gen/                # Generator CLI & wizard interaktif
│   ├── lint/               # Architecture Linter CLI (AST Analyzer)
│   ├── doctor/             # Environment Doctor CLI
│   ├── migrate/            # CLI runner database migration
│   ├── new/                # CLI Scaffolding project baru (gokit-new)
│   ├── rm/                 # CLI penilai/penghapus modul
│   └── seed/               # CLI runner database seeder
├── db/
│   ├── factory/            # Data factories (Gofakeit)
│   ├── migrations/         # File SQL migration (.up.sql & .down.sql)
│   └── seed/               # Database seeders
├── manifests/              # Definisi YAML manifest modul
├── internal/
│   ├── config/             # Modular configuration (http, database, viper, logger, storage)
│   ├── foundation/         # Core framework (appcontext, database, exception, health, logger, mapper, response, validator)
│   ├── middleware/         # Middleware HTTP (auth, request context, correlation id)
│   └── module/             # Domain modules (User, Product, Order, dll)
├── docker-compose.yml      # Local dev setup (MySQL 8.0 & Redis 7)
├── env.example.json        # Template konfigurasi environment
├── Taskfile.yml            # Task runner automation
└── go.mod
```

---

## 🔨 Semua Perintah Task Runner

```bash
# Development & Quality
task dev               # Jalankan server dev dengan hot reload (Air)
task build             # Compile binary release ke bin/app
task test              # Jalankan unit tests + coverage report
task lint              # Jalankan linter checker (golangci-lint)
task lint-arch         # 🔥 Jalankan Architecture Linter (rel lint)
task doctor            # 🩺 Jalankan diagnostik kesehatan lingkungan dev (rel doctor)
task fmt               # Format kode Go
task docs              # Generate OpenAPI / Swagger docs

# Code Generation & Cleanup
task gen name=product  # Generate module CRUD + Migration + Seeder + Factory
task rm name=Product   # Hapus module Product secara permanen
task make-migration    # Buat file SQL migration baru
task make-seeder       # Generate file Seeder
task make-factory      # Generate file Factory

# Database & Migration
task migrate-up        # Terapkan semua skema migration baru
task migrate-down      # Rollback 1 file migration terakhir
task migrate-fresh     # Reset DB, re-apply migrasi dari awal & seed
task seed              # Isi database dengan data seeder awal
```

---
**RelGo Framework** — *Architected for Consistency, Scale, and Joy.*