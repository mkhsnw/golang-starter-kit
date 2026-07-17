# 🚀 Golang Starter Kit (Fiber + GORM + Clean Architecture)

Sebuah **kit murni**, bukan aplikasi contoh — repo ini berisi mesin (generator, middleware, config, auth) yang siap dipakai untuk bootstrap project backend baru dalam hitungan detik, bukan template yang harus kamu fork lalu bersihkan manual.

Dokumen ini menjelaskan **apa isinya, kenapa disusun begini, dan cara pakainya dari nol sampai jalan** — bukan cuma daftar perintah.

---

## 📖 Daftar Isi

1. [Apa yang Kamu Dapat](#-apa-yang-kamu-dapat)
2. [Arsitektur & Alur Request](#️-arsitektur--alur-request)
3. [Bootstrap Project Baru](#-bootstrap-project-baru)
4. [Persiapan Awal (Kit Ini Sendiri)](#️-persiapan-awal-kit-ini-sendiri)
5. [Menjalankan Server](#-menjalankan-server)
6. [Migrasi Database](#-migrasi-database)
7. [Membuat Fitur Baru (Generator)](#️-membuat-fitur-baru-generator)
8. [Kapan Perlu Row-Locking?](#-kapan-perlu-row-locking)
9. [Alur Autentikasi](#-alur-autentikasi)
10. [Format Response API](#-format-response-api)
11. [Testing & Mocks](#-testing--mocks)
12. [Swagger Docs](#-swagger-docs)
13. [Semua Perintah Task](#-semua-perintah-task)
14. [Catatan & Rencana Pengembangan](#-catatan--rencana-pengembangan)

---

## 📦 Apa yang Kamu Dapat

Begitu kamu bootstrap project baru dari kit ini, kamu langsung punya:

- ✅ **Autentikasi lengkap** — register, login, JWT access token + refresh token (bisa di-revoke), logout
- ✅ **Generator CRUD** — 1 command bikin entity, model, repository, usecase, controller, migration, sekaligus auto-registrasi ke DI & route
- ✅ **Row-locking siap pakai** — buat kasus race-condition sensitif (stok, saldo, kuota)
- ✅ **Rate limiting via Redis** — akurat walau nanti di-scale ke banyak instance
- ✅ **Error handling & response format konsisten** — terstandar di semua endpoint tanpa perlu dipikir ulang tiap fitur
- ✅ **Swagger docs otomatis, testing dengan mock, hot-reload development**

Yang **tidak** ikut terbawa ke project baru: modul contoh apa pun. Kit ini sengaja tidak punya "aplikasi contoh" (seperti `Product`/`Order`) di dalamnya — supaya project barumu benar-benar bersih dari hari pertama, tidak perlu menghapus kode yang tidak relevan dengan bisnismu.

---

## 🏛️ Arsitektur & Alur Request

Setiap fitur mengikuti **Clean Architecture** — dipecah ke beberapa layer yang masing-masing punya 1 tanggung jawab. Ini supaya business logic-mu tidak menempel ke framework HTTP atau ke cara kerja database tertentu.

```
HTTP Request
    │
    ▼
┌─────────────┐   Terima request, teruskan ke Controller
│   Route     │
└─────┬───────┘
      ▼
┌─────────────┐   Parsing & validasi input, panggil Usecase,
│  Controller │   bungkus hasil jadi response JSON standar
└─────┬───────┘
      ▼
┌─────────────┐   Business logic hidup di sini — aturan bisnis,
│   Usecase   │   transaksi, locking, orkestrasi antar repository
└─────┬───────┘
      ▼
┌─────────────┐   Akses data mentah (Create/FindByID/Update/Delete)
│ Repository  │
└─────┬───────┘
      ▼
┌─────────────┐
│   Entity    │   Representasi tabel database (struct GORM)
└─────────────┘
      │
      ▼
   Database
```

### Kenapa 1 fitur jadi banyak file?

| File | Isinya | Kenapa dipisah |
|---|---|---|
| `entity/xxx_entity.go` | Struct yang merepresentasikan tabel database | Struktur database terpisah dari struktur API |
| `model/xxx_model.go` | DTO request & response | Field sensitif (misal `Password`) tidak pernah bocor ke JSON response tanpa sengaja |
| `repository/xxx_repository.go` | Query ke database (embed generic `Repository[T]`) | Usecase tidak perlu tahu detail SQL/GORM |
| `usecase/xxx_usecase.go` | Business logic — validasi, aturan bisnis | "Otak" fitur, sengaja dipisah dari HTTP supaya gampang di-unit-test tanpa server jalan |
| `delivery/http/controller/xxx_controller.go` | Terima HTTP request, panggil usecase | Satu-satunya layer yang "tahu" soal Fiber |

---

## 🌱 Bootstrap Project Baru

Ini titik masuk utama kalau kamu mau mulai project baru dari kit ini:

```powershell
task gokit-new github.com/usernamekamu/nama-project-baru
```

Yang terjadi otomatis:
1. Seluruh core (config, middleware, auth, generator) di-copy ke folder sejajar (`../nama-project-baru`)
2. `.git`, `.github`, `env.json`, `bin/`, `docs/` **tidak ikut ter-copy** — project baru mulai dengan git history bersih dan tanpa kebawa secret kit ini
3. Nama module Go otomatis di-rename di semua file (`go.mod`, import path, dst)
4. `go mod tidy` otomatis dijalankan

Setelah itu, masuk ke folder project baru dan lanjut ke [Persiapan Awal](#️-persiapan-awal-kit-ini-sendiri) di bawah — semua langkah selanjutnya sama persis.

---

## 🛠️ Persiapan Awal (Kit Ini Sendiri)

### Instalasi

1. Pastikan Go (minimal versi 1.21) sudah terpasang.
2. Install **Task** — pengganti `make` yang cross-platform:
   ```powershell
   go install github.com/go-task/task/v3/cmd/task@latest
   ```

### Konfigurasi Environment

```powershell
cp env.example.json env.json
```

Isi `env.json` sesuai environment kamu:

```json
{
  "app": { "name": "...", "environment": "dev", "port": 3000, "url": "http://localhost:3000" },
  "database": {
    "host": "localhost", "port": 3306, "username": "root", "password": "", "name": "starterkit",
    "pool": { "maxIdle": 10, "maxOpen": 100, "maxLifetime": "1h" }
  },
  "jwt": {
    "secret": "ganti-dengan-secret-kuat",
    "expiration_hours": 1,
    "refresh_secret": "ganti-dengan-secret-lain-yang-berbeda",
    "refresh_expiration_days": 7
  },
  "redis": { "host": "127.0.0.1", "port": 6379, "password": "", "database": 0 }
}
```

> ⚠️ `env.json` sudah masuk `.gitignore` — jangan pernah commit versi yang isinya kredensial asli. `env.example.json` yang jadi acuan publik.

**Redis wajib jalan** sebelum server di-start — dipakai untuk rate limiting yang konsisten walau nanti aplikasi di-scale ke banyak instance.

### Install Tooling Developer

Sekali di awal, install semua tool pendukung (linter, hot-reload, mock generator, swagger generator, migration CLI):

```powershell
task init
```

> **Windows:** pastikan `C:\Users\<NAMA_USER>\go\bin` sudah masuk `PATH`.

---

## ⚡ Menjalankan Server

Pastikan MySQL & Redis aktif, dan database yang disebut di `env.json` sudah dibuat (database-nya sendiri — tabel dibuat lewat migration).

```powershell
task dev           # jalan dengan hot-reload (Air) — dipakai sehari-hari
go run cmd/main.go # atau jalan langsung tanpa hot-reload
```

Server aktif di `http://localhost:3000`.

---

## 🗄️ Migrasi Database

Pakai migration berbasis file SQL (`golang-migrate`), **bukan** `AutoMigrate()` GORM — supaya perubahan skema database selalu tercatat riwayatnya, dan production tidak pernah mengubah skema secara diam-diam saat aplikasi start.

```powershell
task migrate-up                # terapkan semua migration yang belum jalan
task migrate-down               # rollback 1 migration terakhir
task migrate-version            # lihat versi migration saat ini
task migrate-create name=xxx    # bikin file migration kosong (misal nambah kolom)
```

Jalankan `task migrate-up` sekarang untuk membuat tabel `users` dan `refresh_tokens` yang sudah tersedia bawaan kit ini.

Setiap kali kamu generate modul baru (lihat bawah), file migration otomatis dibuat — tinggal `task migrate-up`, atau tambahkan `migrate=true` supaya langsung diterapkan.

---

## 🏗️ Membuat Fitur Baru (Generator)

Generate 1 modul CRUD lengkap — entity, model, repository, usecase, controller, migration, plus registrasi otomatis ke route & dependency injection — dengan 1 command.

```powershell
task gen name=Product fields="name:string,price:float64,is_active:bool"
```

Semua primary key default pakai **UUID v7**: tetap *time-ordered* (performa index database tetap bagus) tapi tidak bisa ditebak urutannya — mencegah orang luar meng-enumerasi data hanya dengan mengganti angka di URL.

### Foreign Key otomatis

Field berakhiran `_id` otomatis dikenali sebagai foreign key — tipe kolom disesuaikan, index dibuat, dan constraint `ON DELETE CASCADE` ditambahkan otomatis:

```powershell
task gen name=Order fields="user_id:string,total:float64,note:text?" tx=true test=true migrate=true
```

### Semua opsi generator

| Flag | Kegunaan |
|---|---|
| `fields="..."` | Definisi field, format `nama:tipe` atau `nama:tipe?` (nullable) |
| `tx=true` | Bungkus operasi write (Create/Update/Delete) ke dalam database transaction. Pakai kalau modul ini melibatkan >1 tabel yang harus berhasil/gagal bersamaan |
| `test=true` | Generate juga file test (`usecase_test.go`, `controller_test.go`). **Otomatis aktif kalau `tx=true`** — logic transaksional adalah yang paling butuh diuji |
| `dry=true` | Preview file apa saja yang akan dibuat, tanpa menulis apa pun ke disk |
| `force=true` | Timpa file yang sudah ada tanpa konfirmasi |
| `migrate=true` | Langsung jalankan migration setelah generate, tanpa command terpisah |

```powershell
task rm name=Product     # hapus modul + bersihkan semua registrasinya (kebalikan dari gen)
```

---

## 🔒 Kapan Perlu Row-Locking?

Untuk operasi "baca nilai → putuskan sesuatu → tulis ulang nilai itu" (mengurangi stok, memotong saldo), transaksi biasa (`tx=true`) **saja belum cukup** — dua request bersamaan tetap bisa sama-sama membaca nilai lama sebelum salah satu sempat menyimpan perubahan.

`Repository[T]` menyediakan `FindByIDForUpdate` yang mengunci baris tersebut sampai transaksi selesai:

```go
err := u.TxManager.Run(ctx, func(ctxTx context.Context) error {
    product, err := u.ProductRepository.FindByIDForUpdate(ctxTx, req.ProductId)
    if err != nil {
        return exception.NotFound("Product not found")
    }
    if product.Stock < req.Amount {
        return exception.Conflict("Insufficient stock")
    }
    product.Stock -= req.Amount
    return u.ProductRepository.Update(ctxTx, product)
})
```

**Aturan praktis:** kalau cuma menampilkan data (`GetByID` untuk halaman detail), pakai `FindByID` biasa. Locking cuma diperlukan kalau race condition-nya benar-benar bisa merusak data — memakainya di semua tempat hanya akan memperlambat request lain tanpa manfaat.

---

## 🔑 Alur Autentikasi

Modul `User` + autentikasi sudah tersedia bawaan (bukan hasil generate, ditulis manual karena logic-nya berbeda dari CRUD biasa):

```
1. POST /api/v1/auth/register   { name, email, password }
2. POST /api/v1/auth/login      { email, password }
                                 → { "data": { "token": "...", "refresh_token": "..." } }
3. Sertakan access token di setiap request ke endpoint terproteksi:
   Authorization: Bearer <token>

4. Kalau access token kedaluwarsa (default 1 jam):
   POST /api/v1/auth/refresh     { "refresh_token": "..." }
                                 → dapat access token baru

5. POST /api/v1/auth/logout (butuh login) → revoke semua refresh token milik user ini
```

Refresh token disimpan di database dalam bentuk **hash** (bukan mentah) — kalau database bocor, isi kolom itu tidak bisa langsung dipakai untuk login sebagai user manapun.

Endpoint lain yang tersedia: `GET /api/v1/users/current` (butuh login) — mengembalikan data user yang sedang login.

---

## 📦 Format Response API

**Sukses (single item):**
```json
{ "data": { "id": "...", "name": "..." } }
```

**Sukses (list dengan pagination):**
```json
{
  "data": [ { "id": "...", "name": "..." } ],
  "paging": { "page": 1, "size": 10, "total_item": 42, "total_page": 5 }
}
```

**Error:**
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input",
    "fields": [ { "field": "email", "message": "failed on 'required' validation" } ]
  }
}
```

`code` bersifat machine-readable, dipakai frontend untuk `switch-case` tanpa parsing teks. Kode yang tersedia: `VALIDATION_ERROR`, `NOT_FOUND`, `CONFLICT`, `UNAUTHORIZED`, `FORBIDDEN`, `BAD_REQUEST`, `INTERNAL_SERVER_ERROR`.

---

## 🧪 Testing & Mocks

Semua Usecase bergantung ke *interface*, bukan struct konkret — supaya bisa diuji tanpa koneksi database asli.

```powershell
task mock    # generate ulang mock setiap kali interfaces.go berubah
task test    # jalankan semua test + laporan coverage
```

---

## 📖 Swagger Docs

```powershell
task docs
```

Buka **http://localhost:3000/api/v1/docs/index.html** saat server jalan.

> **Catatan:** `swag` (tool generator Swagger) punya dukungan yang belum sepenuhnya matang untuk Go generics. Kalau skema untuk response yang memakai `model.WebResponse[T]` tampil tidak lengkap di dokumentasi, ini keterbatasan tooling pihak ketiga, bukan bug di endpoint-nya.

---

## 🔨 Semua Perintah Task

```powershell
task help              # dokumentasi interaktif semua command
task dev                # jalankan server dengan hot-reload
task build              # compile binary release ke bin/app
task test               # jalankan test + coverage
task lint               # cek code-style (golangci-lint)
task fmt                # format kode
task mock               # generate mock untuk testing
task docs               # generate dokumentasi swagger
task gen name=... fields="..." [tx=true] [test=true] [dry=true] [force=true] [migrate=true]
task rm name=...
task gokit-new <module-path>   # bootstrap project baru dari kit ini
task migrate-up / migrate-down / migrate-version / migrate-create name=xxx
```

---

## 📝 Catatan & Rencana Pengembangan

Supaya ekspektasinya jelas:

- **CI/CD (GitHub Actions) belum ada — ini disengaja, bukan kelewat.** Selama kit ini masih 1 repo yang kamu kembangkan sendiri (belum di-reuse ke banyak project via `gokit-new`), CI belum krusial. **Begitu kamu mulai reuse kit ini ke project kedua**, CI wajib jadi prioritas — supaya perubahan di core tidak diam-diam menyebarkan bug ke semua project yang bergantung ke sini.
- **Refresh token belum dirotasi tiap dipakai** — token yang sama tetap valid sampai expired 7 hari, bukan diganti tiap kali di-refresh. Level lebih aman (rotasi + deteksi reuse) bisa ditambahkan nanti kalau dibutuhkan, tapi implementasi sekarang (revoke saat logout, hash tersimpan, TTL pendek untuk access token) sudah jauh lebih aman dari single-token biasa.
- **Rate limiter di-skip untuk `/docs`** — supaya dokumentasi Swagger tetap bisa diakses tanpa terkena limit endpoint umum.

Kalau kamu berkontribusi ke kit ini, cek 2 poin pertama di atas sebagai starting point paling berdampak.