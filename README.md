# 🚀 Golang Starter Kit (Fiber + GORM + Clean Architecture)

Starter Kit modern untuk pengembangan backend Golang menggunakan **Clean Architecture** (berbasis *Dependency Inversion/Interface*). Proyek ini dilengkapi dengan modul-modul generator (CRUD otomatis), dukungan Swagger UI (OpenAPI), struktur Unit Testing (Testify + Mockery), dan pengelolaan environment/konfigurasi otomatis.

---

## ✨ Fitur Unggulan
- **Fiber v3** sebagai *Web Framework* yang super cepat.
- **GORM** sebagai ORM.
- **Clean Architecture** yang ketat dengan ekstensi **Generic Repository**.
- **Interface-based Usecase/Controller** yang siap pakai untuk *Mocking* di Unit Test.
- **Anotasi Swagger** ter-generasi secara otomatis.
- **Code Generator Tool** bawaan untuk men-generate semua *boilerplate* secara instan.
- **Task Runner (go-task)** - Pengganti Makefile yang cross-platform dan bekerja sempurna di Windows.

---

## 🛠️ 1. Persiapan Awal (Setup)

### Kloning & Instalasi
1. Kloning repositori ini.
2. Pastikan Go (minimal 1.21) terpasang di mesin Anda.
3. Anda tidak butuh `make`. Install **Task runner** sebagai pendamping Anda:
   ```powershell
   go install github.com/go-task/task/v3/cmd/task@latest
   ```

### Konfigurasi Environment (env.json)
Kredensial disimpan dalam `env.json` yang di-ignore oleh git untuk menjaga keamanan.
1. Salin *template* konfigurasi:
   ```powershell
   cp env.example.json env.json
   ```
2. Buka `env.json` dan ubah konfigurasi **Database** (MySQL) serta **JWT Secret** sesuai keinginan Anda.

### Menginstal Dependensi Developer (DX)
Untuk pertama kali, instal *tooling* pendukung (*linter*, *hot reload*, *mockery*, dan *swaggo*):
```powershell
task init
```
> **Penting (Windows):** Pastikan direktori `C:\Users\<NAMA_USER>\go\bin` sudah terdaftar ke dalam Environment Variable `PATH` komputer Anda agar eksekusi `task`, `swag`, dan `mockery` berjalan normal.

---

## ⚡ 2. Menjalankan Server & API

Pastikan server MySQL Anda sedang aktif dan *database* yang disebut dalam `env.json` sudah dibuat.

```powershell
# Menjalankan server dengan live-reload (Air)
task dev

# Atau menjalankan langsung dengan go run
go run cmd/main.go
```
API sekarang akan merespon pada port `3000`!

---

## 🏗️ 3. Membuat Fitur (Generate Module)

Salah satu keunggulan boilerplate ini adalah *Code Generator*. Anda tidak perlu lagi membuat Controller, Usecase, Repository, Model, atau Unit Test satu per satu.

Gunakan CLI generator bawaan (via Task runner). *Perhatikan penggunaan tanda `--` sebelum menyebutkan nama modul*:
```powershell
# Contoh: Membuat fitur Transaction standar
task gen -- Transaction

# Contoh: Membuat fitur Category dengan field spesifik
task gen -- Category --fields name:string,description:string?,is_active:bool
```
Langkah ini secara **otomatis** membuat:
- `internal/entity/transaction.go`
- `internal/model/transaction_model.go`
- `internal/repository/transaction_repository.go`
- `internal/usecase/transaction_usecase.go`
- `internal/delivery/http/controller/transaction_controller.go`
- `internal/usecase/transaction_usecase_test.go` (Unit Test Lengkap 100% Mocks)
- `internal/delivery/http/controller/transaction_controller_test.go` (Integration Test Lengkap)
- Registrasi route & dependency injection baru di `app.go`, `route.go`, dan `interfaces.go`.

**Menghapus Modul (Rollback):**
Jika Anda salah mengetik nama atau sekadar ingin menghapus modul yang baru digenerate secara bersih (termasuk mencabut injeksinya), gunakan perintah `rm`:
```powershell
task rm -- Transaction
```

**Langkah Manual Lanjutan:**
Setelah fitur digenerate, buka `internal/config/gorm.go` lalu tambahkan struct entity yang baru terbentuk (`entity.Transaction{}`) ke fungsi `AutoMigrate()` agar tabel SQL terbentuk otomatis.

---

## 🧪 4. Meng-generate Mocks & Unit Test

Semua Usecase di proyek ini terpisah dari database oleh *Interface*. Setiap Anda meng-generate modul baru, atau mengedit fungsi Interface (`interfaces.go`), Anda perlu memperbarui **Mock**. Generator secara otomatis men-generate file Test yang komprehensif, Anda hanya perlu menjalankan:

```powershell
# Perbarui mock file untuk Test
task mock
```
Setelah Mock dibuat, file test yang dihasilkan (sudah dilengkapi dengan `testify/assert` dan `testify/mock`) siap dikompilasi!

Untuk menjalankan seluruh Test sekaligus melihat **Laporan Cakupan (Coverage Report)**:
```powershell
task test
```

---

## 📖 5. Meng-generate Swagger Docs

Anotasi Swagger telah disisipkan ke seluruh Controller saat *Code Generator* dieksekusi. Jika Anda menambah endpoint atau mengubah model parameter, perbarui dokumen API dengan:

```powershell
# Men-generate ulang docs
task docs
```
Lalu kunjungi saat server menyala:
👉 **[http://localhost:3000/docs/index.html](http://localhost:3000/docs/index.html)**

---

## 🔨 Perintah Lainnya

Boilerplate menyediakan script `Taskfile.yml` singkat untuk kebutuhan esensial:
- `task lint` : Mengecek *code-style* menggunakan `golangci-lint`
- `task fmt`  : Memformat ulang kode agar rapi (`go fmt`)
- `task build`: Meng-compile *binary release* aplikasi ke direktori `bin/app`.
