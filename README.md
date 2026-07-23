# 🚂 RelGo Framework (v1.0.3)

> **RelGo — Framework Go opini yang bikin kamu ngoding REST API ngebut, rapi, dan anti-ribet, mulus kayak kereta berjalan di atas rel.**

---

## ⚡ Quickstart (5 Menit Langsung Running!)

Gak usah ribet setup boilerplate dari nol. Ikuti alur cepat ini untuk mulai ngoding:

### 1. Install CLI `rel` secara Global
Buka terminal kamu (di mana aja) dan jalankan:
```bash
go install github.com/mkhsnw/rel/cmd/rel@latest
```

### 2. Bikin Proyek Baru
Pindah ke folder workspace kamu, lalu buat proyek baru:
```bash
rel new github.com/username/my-awesome-api
```
*(CLI bakal otomatis mengunduh template, setup struktur folder, dan merename module Go kamu)*.

### 3. Masuk ke Folder Proyek & Setup Environment
```bash
cd my-awesome-api

# Copy file konfigurasi environment
cp env.example.json env.json

# Jalankan MySQL 8 & Redis 7 lokal via Docker
docker-compose up -d
```

### 4. Cek Kesehatan Setup Kamu (`rel doctor`)
Jalankan diagnostik untuk memastikan semua tool & koneksi database udah aman:
```bash
rel doctor
```

### 5. Generate Modul CRUD Pertama Kamu (misal: Product)
```bash
rel gen Product
```
*(Otomatis bikin Controller, Service, Repository, DTO, Route, Migration SQL, Seeder, dan Factory)*.

### 6. Jalankan Migrasi Database
```bash
rel migrate up
```

### 7. Jalankan Server Dev dengan Live Reload!
```bash
task dev
# atau manual: go run cmd/main.go
```
🔥 Server REST API kamu sekarang sudah jalan di `http://localhost:3000` dengan Swagger UI di `http://localhost:3000/swagger/index.html`!

---

## 🎯 Filosofi "3-Touchpoints Contract"

Di RelGo, kamu gak akan pusing ngurusin boilerplate routing, response wrapper, DTO mapping, atau error catalog berulang-ulang.

Untuk nambah fitur atau modul baru, kamu **cuma perlu nyentuh 3 area**:
1. **Domain Entity** (`internal/module/<name>/entity.go`) ➔ Definisi struct & tabel DB.
2. **Repository Query** (`internal/module/<name>/repository.go`) ➔ Query khusus jika butuh SQL kompleks.
3. **Business Service** (`internal/module/<name>/service.go`) ➔ Aturan bisnis utama aplikasi kamu.

Sisa hal printilan seperti HTTP Routing, JSON Standard Response, Validasi Payload, Context Propagation, Logging, dan OpenAPI/Swagger **otomatis ditangani oleh Foundation & CLI RelGo**.

---

## 🛠️ Penyesuaian & Kustomisasi Kode (Flexibility First)

> 💡 **Catatan Hasil Generator (`rel gen`):**
> Kode yang di-generate oleh `rel gen` disiapkan sebagai **boilerplate fondasi awal (80–90%)** berdasarkan manifest/entity yang kamu buat.
> 
> Kamu **bebas dan sangat disarankan** untuk melakukan penyesuaian (kustomisasi) kapan saja:
> - **Struct Entity** (`entity.go`): Tambahkan relasi database GORM (`has_many`, `belongs_to`), indeks, atau field tambahan.
> - **DTO Payload & Response** (`dto/`): Ubah format payload request/response JSON agar pas dengan kebutuhan API/Mobile kamu.
> - **Repository Query** (`repository.go`): Tulis query SQL khusus, JOIN multi-tabel, atau agregasi.
> - **Service Logic** (`service.go`): Masukkan aturan bisnis spesifik dari aplikasi kamu.

---

## 🧰 Perintah CLI `rel` Cheat Sheet

| Perintah | Fungsi |
| :--- | :--- |
| `rel new <module-path>` | Bikin proyek RelGo baru dari awal |
| `rel gen [module-name]` | Generate modul CRUD / Bisnis (ada mode wizard interaktif) |
| `rel rm <module-name>` | Hapus modul yang udah di-generate secara bersih |
| `rel migrate up` | Jalankan semua migrasi database SQL yang belum dieksekusi |
| `rel migrate down` | Rollback 1 langkah migrasi database |
| `rel migrate fresh` | Drop semua tabel & jalankan ulang migrasi dari nol |
| `rel seed [count=50]` | Isi database dengan data seeder awal |
| `rel make-migration <name>` | Bikin file SQL migration baru (`.up.sql` & `.down.sql`) |
| `rel doctor` | Diagnostik kesehatan environment, Redis, & MySQL kamu |
| `rel lint` | Cek aturan arsitektur (AST Linter) agar kode gak acak-acakan |

---

## 📊 Compatibility Matrix

| Komponen | Versi Teruji | Keterangan |
| :--- | :--- | :--- |
| **Go Compiler** | `Go 1.25.0+` | Minimal versi Go yang dibutuhkan |
| **Web Engine** | `Fiber v3 (v3.4.0)` | High-performance Fasthttp web engine |
| **ORM / Database** | `GORM v1.31+` / `MySQL 8.0` | Driver MySQL & MariaDB |
| **Cache & Storage** | `Redis v7.0` | Rate limiter & session storage |
| **Config Loader** | `Viper v1.21` | Fail-fast JSON & Env validation |