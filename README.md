# 🚀 Golang Starter Kit (Fiber + GORM + Clean Architecture)

Starter kit backend Go yang dirancang supaya kamu (atau siapa pun yang lanjutin project ini) bisa langsung fokus nulis business logic, bukan sibuk bikin ulang boilerplate CRUD, wiring dependency injection, atau mikirin format error yang konsisten.

Dokumen ini nggak cuma ngasih tahu perintah apa yang harus diketik, tapi juga **kenapa** struktur project ini dibuat begini, dan **bagaimana** semua bagiannya saling terhubung.

---

## 📖 Daftar Isi

1. [Arsitektur & Alur Request](#-arsitektur--alur-request)
2. [Persiapan Awal](#️-persiapan-awal-setup)
3. [Menjalankan Server](#-menjalankan-server)
4. [Migrasi Database](#-migrasi-database)
5. [Membuat Fitur Baru (Generator)](#️-membuat-fitur-baru-generator)
6. [Kapan Perlu Row-Locking?](#-kapan-perlu-row-locking)
7. [Alur Autentikasi](#-alur-autentikasi)
8. [Format Response API](#-format-response-api)
9. [Testing & Mocks](#-testing--mocks)
10. [Swagger Docs](#-swagger-docs)
11. [Perintah Lainnya](#-perintah-lainnya)
12. [Catatan & Rencana Pengembangan](#-catatan--rencana-pengembangan)

---

## 🏛️ Arsitektur & Alur Request

Setiap fitur di project ini mengikuti **Clean Architecture** — dipecah jadi beberapa layer yang masing-masing punya 1 tanggung jawab spesifik. Ini sengaja dipisah supaya business logic kamu nggak nempel ke framework HTTP atau ke cara kerja database — kalau suatu hari kamu ganti Fiber ke framework lain, atau ganti MySQL ke Postgres, layer di tengah (`usecase`) idealnya nggak perlu disentuh sama sekali.

Begini alur 1 request dari masuk sampai keluar:

```
HTTP Request
    │
    ▼
┌─────────────┐   Terima request, parsing & validasi input,
│   Route     │   panggil Controller yang sesuai
└─────┬───────┘
      ▼
┌─────────────┐   Terima input yang sudah tervalidasi,
│  Controller │   panggil Usecase, bungkus hasilnya jadi
└─────┬───────┘   response JSON standar
      ▼
┌─────────────┐   Tempat business logic & aturan bisnis
│   Usecase   │   hidup (cek duplikat, hitung stok, dst)
└─────┬───────┘
      ▼
┌─────────────┐   Akses data mentah ke database
│ Repository  │   (Create/FindByID/Update/Delete/dst)
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

Kalau kamu generate 1 modul (misal `Product`), akan muncul beberapa file berbeda — ini bukan biar ribet, tiap file punya tugas jelas:

| File | Isinya | Kenapa dipisah |
|---|---|---|
| `entity/product_entity.go` | Struct yang merepresentasikan tabel `products` di database | Supaya struktur database terpisah dari struktur API |
| `model/product_model.go` | DTO request (`CreateProductRequest`) & response (`ProductResponse`) | Supaya field sensitif (misal `Password` di `User`) nggak pernah kebawa ke JSON response tanpa sengaja |
| `repository/product_repository.go` | Query ke database (pakai generic `Repository[T]` di baliknya) | Supaya usecase nggak perlu tahu detail SQL/GORM |
| `usecase/product_usecase.go` | Business logic — validasi, aturan bisnis, transaksi | Ini "otak" fitur, dan sengaja dipisah dari HTTP supaya gampang di-unit-test tanpa perlu jalanin server |
| `delivery/http/controller/product_controller.go` | Terima HTTP request, panggil usecase, kembalikan response | Satu-satunya layer yang "tahu" soal Fiber |

---

## 🛠️ Persiapan Awal (Setup)

### Kloning & Instalasi

1. Kloning repository ini.
2. Pastikan Go (minimal versi 1.21) sudah terpasang.
3. Install **Task** — pengganti `make` yang cross-platform, jalan normal di Windows:
   ```powershell
   go install github.com/go-task/task/v3/cmd/task@latest
   ```

### Konfigurasi Environment (`env.json`)

Kredensial disimpan di `env.json`, yang sudah otomatis di-*ignore* git supaya nggak ke-commit nggak sengaja.

```powershell
cp env.example.json env.json
```

Buka `env.json`, sesuaikan bagian `database` (kredensial MySQL) dan `jwt` (secret key). **Jangan pernah commit `env.json` yang isinya kredensial asli** — itu kenapa file ini digitignore, `env.example.json` yang jadi acuan untuk orang lain.

### Install Tooling Developer

Sekali di awal, install semua tool pendukung (linter, hot-reload, mock generator, swagger generator):

```powershell
task init
```

> **Windows:** pastikan `C:\Users\<NAMA_USER>\go\bin` sudah masuk `PATH`, supaya `task`, `swag`, dan `mockery` bisa dipanggil dari terminal mana pun.

---

## ⚡ Menjalankan Server

Pastikan MySQL sudah aktif dan database yang disebut di `env.json` sudah dibuat (database-nya sendiri, bukan tabelnya — tabel dibuat lewat migration, lihat section berikutnya).

```powershell
task dev          # jalan dengan hot-reload (Air) — dipakai sehari-hari
go run cmd/main.go # atau jalan langsung tanpa hot-reload
```

Server aktif di `http://localhost:3000`.

---

## 🗄️ Migrasi Database

Project ini pakai migration berbasis file SQL (`golang-migrate`), **bukan** `AutoMigrate()` GORM — supaya perubahan skema database selalu tercatat jelas riwayatnya, dan supaya production nggak pernah ubah skema secara "diam-diam" saat aplikasi start.

```powershell
task migrate-up              # terapkan semua migration yang belum jalan
task migrate-down             # rollback 1 migration terakhir
task migrate-version          # lihat versi migration saat ini di database
task migrate-create name=xxx  # bikin file migration kosong (misal nambah kolom ke tabel yang sudah ada)
```

Setiap kali kamu generate modul baru lewat `task gen` (lihat section berikutnya), file migration `.up.sql`/`.down.sql` otomatis dibuat di `db/migration/` — tinggal `task migrate-up` buat menerapkannya, atau tambahkan flag `migrate=true` supaya langsung diterapkan otomatis.

---

## 🏗️ Membuat Fitur Baru (Generator)

Ini fitur andalan starter kit ini: generate 1 modul CRUD lengkap (entity, model, repository, usecase, controller, test, migration, plus registrasi otomatis ke route & dependency injection) hanya dengan 1 command.

```powershell
task gen name=Product fields="name:string,price:float64,is_active:bool"
```

Semua primary key secara default pakai **UUID v7** (bukan angka auto-increment biasa). Ini keputusan sadar: UUID v7 tetap *time-ordered* (jadi performa index database tetap bagus), tapi ID-nya nggak bisa ditebak urutannya — mencegah orang luar nge-enumerasi data kamu cuma dengan mengganti angka di URL (`/products/1`, `/products/2`, dst).

### Field dengan Foreign Key

Field yang namanya berakhiran `_id` (misal `category_id`) otomatis dikenali sebagai foreign key:

```powershell
task gen name=Order fields="user_id:string,total:float64,note:text?" tx=true migrate=true
```

Generator otomatis akan:
1. Set tipe kolom SQL jadi `VARCHAR(36)` (cocok dengan UUID v7 tabel yang direferensikan)
2. Bikin index (`idx_orders_user_id`)
3. Tambahkan foreign key constraint dengan `ON DELETE CASCADE`

### Kapan pakai flag `tx=true`?

Tambahkan `tx=true` kalau operasi Create/Update/Delete modul ini **melibatkan lebih dari 1 tabel yang harus berhasil/gagal bersamaan** (misal: bikin `Order` sekaligus mengurangi `Stock` di `Product`). Flag ini membungkus operasi write ke dalam database transaction — kalau salah satu langkah gagal di tengah jalan, semua perubahan otomatis dibatalkan (rollback), nggak ada data setengah-jadi yang nyangkut.

Kalau modul kamu berdiri sendiri tanpa keterkaitan ke tabel lain (misal `Category` yang cuma disimpan sendiri), nggak perlu flag ini.

### File yang otomatis dibuat

```
internal/entity/product_entity.go
internal/model/product_model.go
internal/repository/product_repository.go
internal/usecase/product_usecase.go
internal/delivery/http/controller/product_controller.go
internal/usecase/product_usecase_test.go
internal/delivery/http/controller/product_controller_test.go
db/migration/xxxxx_create_products_table.up.sql (+ .down.sql)
```

Plus registrasi otomatis ke `app.go`, `route.go`, dan `interfaces.go` — kamu nggak perlu wiring manual apa pun.

### Opsi lain generator

```powershell
task gen name=Product fields="..." dry=true    # preview dulu, nggak nulis file apa pun
task gen name=Product fields="..." force=true  # timpa file yang sudah ada
task rm name=Product                            # hapus modul + bersihkan semua registrasinya
```

---

## 🔒 Kapan Perlu Row-Locking?

Untuk operasi yang sifatnya "baca nilai → putuskan sesuatu → tulis ulang nilai itu" (misal mengurangi stok, memotong saldo), transaksi biasa (`tx=true`) **saja belum cukup** — dua request yang datang bersamaan tetap bisa sama-sama baca nilai lama sebelum salah satu sempat menyimpan perubahan.

Untuk kasus ini, `Repository[T]` menyediakan `FindByIDForUpdate` yang mengunci baris tersebut sampai transaksi selesai:

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

**Aturan praktis:** kalau cuma menampilkan data (`GetByID` untuk halaman detail), pakai `FindByID` biasa — jangan pakai locking di situ karena cuma akan memperlambat request lain tanpa manfaat. Locking cuma diperlukan kalau race condition-nya benar-benar bisa merusak data (stok minus, saldo dobel terpotong, dsb).

---

## 🔑 Alur Autentikasi

```
1. POST /api/v1/auth/register   → { name, email, password }
2. POST /api/v1/auth/login      → { email, password }
                                 → balikin { "data": { "token": "xxx" } }
3. Sertakan token di setiap request ke endpoint yang butuh login:

   Authorization: Bearer xxx
```

Middleware auth akan otomatis menolak (401) kalau token nggak ada, salah format, atau sudah kedaluwarsa — kamu nggak perlu cek ulang manual di tiap controller.

---

## 📦 Format Response API

Semua response API (sukses maupun error) mengikuti bentuk yang sama, supaya predictable buat siapa pun yang konsumsi API ini.

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
    "fields": [
      { "field": "email", "message": "failed on 'required' validation" }
    ]
  }
}
```

`code` di sini machine-readable — dipakai frontend buat `switch-case` tanpa perlu parsing teks pesan errornya. Beberapa `code` yang tersedia: `VALIDATION_ERROR`, `NOT_FOUND`, `CONFLICT`, `UNAUTHORIZED`, `FORBIDDEN`, `BAD_REQUEST`.

---

## 🧪 Testing & Mocks

Semua Usecase bergantung ke *interface*, bukan struct konkret — supaya bisa diuji tanpa perlu koneksi database asli.

Setiap kali generate modul baru atau mengubah `interfaces.go`, perbarui mock-nya:
```powershell
task mock
```

Lalu jalankan semua test sekaligus lihat laporan cakupannya:
```powershell
task test
```

---

## 📖 Swagger Docs

Anotasi Swagger otomatis tersisip di controller saat generate. Setiap ada perubahan endpoint/model, regenerate dokumentasinya:

```powershell
task docs
```

Lalu buka (saat server jalan): **http://localhost:3000/docs/index.html**

---

## 🔨 Perintah Lainnya

```powershell
task help    # dokumentasi interaktif semua command
task lint    # cek code-style (golangci-lint)
task fmt     # format ulang kode (go fmt)
task build   # compile binary release ke bin/app
```

---

## 📝 Catatan & Rencana Pengembangan

Supaya ekspektasinya jelas — starter kit ini sudah solid untuk mulai membangun dan deploy aplikasi skala kecil-menengah. Beberapa hal berikut **belum** ada dan disengaja untuk ditambahkan nanti kalau kebutuhannya sudah nyata (bukan diantisipasi dari awal):

- **CI/CD** (GitHub Actions) — `task test` & `task lint` belum otomatis jalan tiap push
- **Refresh token** — JWT saat ini single-token, belum bisa di-revoke sebelum expired
- **Rate limiter terdistribusi (Redis)** — limiter saat ini in-memory, akurat selama aplikasi jalan di 1 instance; perlu dipindah ke Redis kalau nanti di-scale ke banyak instance

Kalau kamu berkontribusi ke project ini, silakan cek 3 poin di atas dulu sebagai starting point.