# 🚀 Golang Starter Kit: An Opinionated Framework

Repositori ini bukan lagi sekadar "starter kit" atau kumpulan boilerplate. Ini adalah **opinionated framework** yang dirancang untuk mencegah developer melakukan kesalahan arsitektural. Fokus utamanya adalah **konsistensi**, **kemudahan maintain**, dan **developer experience** yang superior.

## 📖 Daftar Isi

1. [Filosofi Arsitektur (The "Why")](#-filosofi-arsitektur-the-why)
2. [Arsitektur Beku (Frozen Architecture)](#-arsitektur-beku-frozen-architecture)
3. [Generator: Murid dari Foundation](#-generator-murid-dari-foundation)
4. [Apa yang Kamu Dapat](#-apa-yang-kamu-dapat)
5. [Memulai Project Baru](#-memulai-project-baru)
6. [Struktur Konfigurasi](#️-struktur-konfigurasi)
7. [Format Response API](#-format-response-api)
8. [Testing & Mocks](#-testing--mocks)
9. [Semua Perintah Task](#-semua-perintah-task)

---

## 🧠 Filosofi Arsitektur (The "Why")

Sebelum menggunakan framework ini, penting untuk memahami **mengapa** keputusan arsitektur tertentu diambil:

### 1. Kenapa ada `foundation`?
Foundation adalah tulang punggung framework. Alih-alih membuat logic paginasi, format error, atau struktur response secara berulang di setiap fitur, semuanya ditarik ke tengah. `foundation` memastikan bahwa **seluruh API berbicara dengan bahasa yang sama**. Tidak ada lagi module yang me-return error dengan format berbeda.

### 2. Kenapa ada `module`?
Aplikasi sering kali menjadi monolith yang sulit dipecah karena *logic* saling silang. Dengan mengelompokkan kode berdasarkan `module` (domain driven), setiap fitur (seperti User, Order, Product) terisolasi dengan rapi. Jika suatu saat fitur harus dipindah ke microservice, pemisahannya jauh lebih mudah.

### 3. Kenapa DTO dipisah?
Dulu, kita sering mencampur Entity (struktur database) dengan DTO (struktur API). Akibatnya, password atau field sensitif bisa bocor ke response secara tak sengaja. Dengan memisahkan DTO, **Entity hanya urusan database, dan DTO hanya urusan transport**. Tidak ada *Single Source of Truth* yang rancu.

### 4. Kenapa Controller dibuat setipis mungkin (Thin Controller)?
Controller hanya boleh tahu soal framework HTTP (Fiber). Tugasnya hanya parsing request, memanggil Usecase, dan me-return JSON lewat `foundation/response`. Jika suatu saat kita pindah dari Fiber ke Gin atau gRPC, business logic (Usecase) tidak perlu disentuh sama sekali.

### 5. Kenapa Generator sangat beropini (Opinionated Generator)?
Generator yang terlalu fleksibel akan menghasilkan kode yang tidak konsisten. Generator di sini dipaksa patuh pada aturan Foundation. Ia memaksa developer membuat kode dengan standar kualitas yang sama, lengkap dengan DTO yang benar, mapper yang benar, dan error handling yang immutable.

---

## 🧊 Arsitektur Beku (Frozen Architecture)

**Arsitektur pada folder structure ini telah dibekukan.**

Tidak akan ada lagi perdebatan apakah DTO ditaruh di dalam atau di luar module, atau bagaimana file konfigurasi disusun. Semuanya sudah pada bentuk finalnya yang paling matang.

Perubahan ke depan hanya akan fokus pada:
1. **Penyempurnaan Generator**: Membuat generator semakin cerdas membaca *manifest*.
2. **Implementasi Foundation**: Menambahkan fitur-fitur generic yang bisa dipakai semua module.
3. **Bug Fixes**.

Dengan arsitektur yang beku, *Generator* tidak akan pernah tertinggal.

---

## 🤖 Generator: Murid dari Foundation

Generator di framework ini bukanlah sekadar alat copy-paste. Ia adalah "murid" yang patuh pada arsitektur. Generator digerakkan sepenuhnya oleh **Manifest** (file YAML), sehingga satu definisi manifest bisa menghasilkan struktur kode yang 100% konsisten dengan filosofi Foundation.

### 1. Tipe Modul (Standard vs Business)

Framework mendukung 2 tipe modul yang dapat dikonfigurasi melalui manifest:

* **Modul Standard (`type: standard` / Default)**:
  * Digunakan untuk mengelola 1 tabel database (CRUD + Entity + GORM + Migration SQL + Repository + Usecase + Controller + DTO + Route).
  * Field pada `fields` otomatis dipetakan ke **Kolom Migration SQL**, **Struct Entity GORM**, dan **DTO Request/Response**.
* **Modul Bisnis (`type: business`)**:
  * Digunakan untuk **Business Action / Orchestration** (misal: `Checkout`, `PaymentProcess`, `TransferFund`) yang **tidak menguasai 1 tabel database sendiri**.
  * Tidak menghasilkan Entity maupun Migration SQL.
  * Field pada `fields` dipetakan sebagai **payload input & output API (DTO `ProcessRequest` & `ProcessResponse`)**.

### 2. Opsi Transaksi (Transactional Module)

* Set `transactions: true` (atau `transactional: true`) di manifest.
* Pada **Modul Standard**, semua aksi *database write* (`Create`, `Update`, `Delete`) di Usecase/Service otomatis dibungkus di dalam **Database Transaction** (`txManager`).
* Pada **Modul Bisnis**, Service otomatis menerima `s.Tx` (`TransactionManager`) untuk melakukan transaksi *multi-repository* via `s.Tx.RunInTx(ctx, ...)`.

---

### Contoh Manifest YAML

#### A. Standard CRUD Module (`manifests/product.yaml`)
```yaml
name: Product
type: standard
transactions: true
tests: false
fields:
  - name: sku
    type: string
    required: true
  - name: price
    type: float64
    required: true
  - name: published_at
    type: time.Time       # Menggunakan tipe waktu Go (Auto import package "time")
    required: false
  - name: status
    type: string
    sql_type: "ENUM('ACTIVE', 'INACTIVE')"
    required: true
```

#### B. Business Action Module (`manifests/checkout.yaml`)
```yaml
name: Checkout
type: business
transactions: true        # Inject TxManager untuk multi-table transaction
tests: false
fields:
  - name: cart_id
    type: string
    required: true
  - name: payment_method
    type: string
    required: true
  - name: processed_at
    type: time.Time       # Tipe waktu untuk payload DTO request/response
    required: false
```

Eksekusi generator:
```powershell
task gen name=product
```

---

### Tipe Data & SQL Mapping

| Tipe di Manifest (`type`) | Tipe Data Go | Tipe Data SQL (GORM/Migration) | Keterangan |
| :--- | :--- | :--- | :--- |
| `string` / `text` | `string` | `VARCHAR(255)` / `TEXT` | Teks pendek atau panjang. |
| `bool` | `bool` | `TINYINT(1)` | Boolean (true/false). |
| `int`, `int64`, dsb | `int` / `int64` | `INT`, `BIGINT`, dsb | Angka bulat. |
| `float32`, `float64` | `float32`/`64` | `FLOAT` / `DOUBLE PRECISION`| Angka desimal. |
| `time` / `time.Time` | `time.Time` | `TIMESTAMP` | Waktu spesifik. Otomatis mengimport package `"time"`. |

**Aturan Khusus (Smart Generator):**
1. **Custom SQL Type (Enum)**: Gunakan properti `sql_type` (seperti contoh di atas) untuk menimpa tipe SQL bawaan (misal untuk ENUM). Tipe di Go akan tetap menggunakan nilai `type`.
2. **Foreign Key Otomatis**: Jika nama kolom berakhiran `_id` (misal `category_id`), SQL type-nya otomatis di-*override* menjadi `CHAR(36)` (UUID).
3. **Nullable**: Jika `required: false` (atau tidak ditulis), kolom tidak akan memiliki constraint `NOT NULL`.
4. **Field Bawaan (Modul Standard)**: Jangan tulis `id`, `created_at`, dan `updated_at` di manifest modul standard. Ketiga kolom ini wajib dan selalu otomatis di-generate oleh template (menggunakan UUID v7).

---

## 📦 Apa yang Kamu Dapat

- ✅ **Single Source of Truth**: Tidak ada lagi package `model` yang ambigu. Transport diurus DTO, logic oleh Entity, format oleh Foundation.
- ✅ **Config Granular**: `http.go`, `database.go`, `logger.go`, `storage.go`. Generator-friendly dan menghindari konflik *merge*.
- ✅ **Mapper Sentralistik**: Semua error validasi atau meta pagination dipetakan seragam oleh `foundation/mapper`.
- ✅ **Semantic Response**: Response dibentuk dengan makna yang jelas via `response.Created()`, `response.NoContent()`, dsb.
- ✅ **Immutable Exceptions**: Error diinisialisasi sekali (`var ErrUserNotFound = exception.New(...)`) dan kompatibel penuh dengan `errors.Is`.
- ✅ **Row-Locking & Transaksi**: Siap pakai untuk kasus *race-condition*.

---

## 🌱 Memulai Project Baru

```powershell
task gokit-new github.com/usernamekamu/nama-project-baru
```

Yang terjadi:
1. Core framework disalin ke folder baru.
2. Riwayat `.git`, secret di `env.json` tidak ikut terbawa.
3. Nama module Go diotomatisasi ulang.

Setelah itu, copy `env.example.json` menjadi `env.json`, jalankan Redis dan MySQL, lalu:

```powershell
task init
task dev
```

---

## 🗄️ Struktur Konfigurasi

Konfigurasi dipecah menjadi file-file modular di `internal/config`:
- `http.go`: Konfigurasi server Fiber, middleware, error handler.
- `database.go`: Konfigurasi koneksi MySQL via GORM.
- `logger.go`: Logrus dengan pemformatan yang seragam.
- `storage.go`: Redis untuk *rate limiting* dan *caching*.

Pemisahan ini membuat *Generator* dapat menyisipkan konfigurasi baru (seperti Kafka atau AWS S3) tanpa mengganggu konfigurasi yang sudah ada.

---

## 📦 Format Response API

**Semua endpoint wajib menggunakan `foundation/response`**.

Contoh sukses:
```json
{ "data": { "id": "...", "name": "..." } }
```

Contoh error (otomatis di-map dari Validator):
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input",
    "fields": [ { "field": "email", "message": "failed on 'required' validation" } ]
  }
}
```

---

## 🧪 Testing & Mocks

Semua Usecase bergantung ke *interface*, memungkinkannya diuji sepenuhnya tanpa koneksi database.

```powershell
task mock    # generate ulang mock
task test    # jalankan test + coverage
```

---

## 🔨 Semua Perintah Task

```powershell
task help              # List perintah
task dev               # Hot-reload development
task build             # Compile ke binary
task test              # Test + coverage
task lint              # Golangci-lint
task fmt               # Format kode
task mock              # Generate mock
task docs              # Generate swagger
task gen name=...      # Generate CRUD module
task rm name=...       # Hapus module
task migrate-up        # Eksekusi migration
task gokit-new ...     # Buat project baru
```