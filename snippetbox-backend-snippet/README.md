# Snippetbox Backend Snippet

Service snippet Snippetbox: membuat dan membaca snippet yang kedaluwarsa sendiri. Go, Echo v5, GORM, MySQL, JWT.

Service ini tidak menyimpan pengguna dan tidak memanggil service lain. Untuk membuat snippet, token dari service user diverifikasi secara lokal memakai `JWT_SECRET` yang sama, tanpa query ke database akun.

## Struktur

```
cmd/backend/          entry point, hanya merakit dependensi
docs/                 openapi.json dan swagger.html (di-embed ke biner)
internal/
  app/                composition root: menyambung repository, service, handler, router
  config/             membaca environment variable
  database/           koneksi dengan retry, migrasi tabel snippets
  model/              entity GORM dan aturan domain (Snippet, Principal)
  dto/                request, response, dan validasi input
  apperror/           sentinel error dan ValidationError
  validator/          helper validasi
  repository/         akses database (GORM)
  service/            business logic dan TokenAuthenticator; interface didefinisikan di sini
  handler/            lapisan HTTP; interface service didefinisikan di sini
  middleware/         request id, access log, recover, secure headers, CORS, body limit, auth
  router/             definisi route dan rantai middleware
  security/           verifikasi JWT (HS256, issuer snippetbox)
  server/             http.Server dan graceful shutdown
  testutil/           fake, jam tetap, penerbit token uji, helper SQLite untuk test
test/
  integration/        repository, database, service dengan SQLite sungguhan
  e2e/                seluruh aplikasi lewat HTTP
```

Arah dependensi satu arah: `handler -> service -> repository -> model`. Interface didefinisikan oleh pemakainya (`service` mendefinisikan `SnippetRepository` dan `TokenVerifier`, `handler` mendefinisikan `SnippetService`), bukan oleh pembuatnya.

Middleware `Auth` memanggil `service.TokenAuthenticator`, yang memverifikasi token dan menghasilkan `model.Principal` berisi id pengguna. Konsekuensinya, token tetap berlaku sampai kedaluwarsa walaupun akun sudah dihapus di service user.

## Endpoint

| Method | Path | Auth | Hasil |
| --- | --- | --- | --- |
| GET | `/ping` | - | `{"status":"ok"}` |
| GET | `/swagger/` | - | Swagger UI; spesifikasi di `/swagger/openapi.json` |
| GET | `/api/v1/snippets` | - | 10 snippet aktif terbaru |
| GET | `/api/v1/snippets/{id}` | - | snippet aktif |
| POST | `/api/v1/snippets` | Bearer | `201` dengan snippet dan header `Location` |

Route `/api/v1/auth/*` tidak ada di service ini; route itu dimiliki service user.

## Konfigurasi

| Variabel | Default | Keterangan |
| --- | --- | --- |
| `JWT_SECRET` | wajib | minimal 32 karakter; harus sama dengan service user |
| `HTTP_ADDR` | `:8080` | `.env.example` memakai `:8082` agar tidak bentrok dengan service user saat dijalankan di host |
| `APP_ENV` | `development` | |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |
| `DB_DSN` | kosong | jika diisi, menggantikan semua `DB_*` di bawah |
| `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD` | `localhost`, `3306`, `snippetbox`, `snippetbox`, kosong | database milik service ini |
| `DB_CONNECT_RETRIES` | `30` | |
| `DB_CONNECT_RETRY_DELAY` | `2s` | |
| `CORS_ALLOWED_ORIGINS` | `http://localhost:3000` | dipisah koma, `*` untuk semua |

Salin `.env.example` menjadi `.env`; `make` memuatnya otomatis.

## Perintah

```
make run               jalankan server
make build             biner statis di bin/backend
make tidy              go mod tidy
make fmt vet           format dan analisis statis (vet ikut memeriksa file bertag)
make test              unit test
make test/race         unit test dengan race detector
make test/integration  integration test
make test/e2e          end-to-end test
make test/all          semuanya
make test/cover        semuanya dengan cakupan gabungan
```

## Strategi test

| Tingkat | Lokasi | Tag | Yang diuji |
| --- | --- | --- | --- |
| Unit | di samping kode, `*_test.go` | tanpa tag | satu paket, tanpa I/O. Service memakai fake repository dan verifier; handler memakai fake service. |
| Integration | `test/integration` | `-tags=integration` | GORM, migrasi, retry koneksi, dan service + repository + verifikasi JWT pada SQLite in-memory. |
| E2E | `test/e2e` | `-tags=e2e` | `httptest.Server` dengan wiring yang sama dengan produksi, memakai klien HTTP sungguhan dan token yang diterbitkan `testutil.IssueToken`. |

`testutil.IssueToken` menerbitkan token dengan klaim yang sama dengan service user (HS256, issuer `snippetbox`, subject berupa id pengguna). Integration dan e2e memakai SQLite (`glebarez/sqlite`, tanpa cgo), bukan MySQL.
