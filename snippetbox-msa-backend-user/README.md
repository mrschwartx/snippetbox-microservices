# SnippetBox Backend User

Service akun SnippetBox: signup, login, logout, dan data pengguna saat ini. Go, Echo v5, GORM, MySQL, JWT.

Service ini satu-satunya yang menerbitkan token. Service snippet hanya memverifikasi token tersebut memakai `JWT_SECRET` yang sama.

## Struktur

```
cmd/backend/          entry point, hanya merakit dependensi
docs/                 openapi.json dan swagger.html (di-embed ke biner)
internal/
  app/                composition root: menyambung repository, service, handler, router
  config/             membaca environment variable
  database/           koneksi dengan retry, migrasi tabel users
  model/              entity GORM dan aturan domain (User, Token)
  dto/                request, response, dan validasi input
  apperror/           sentinel error dan ValidationError
  validator/          helper validasi
  repository/         akses database (GORM)
  service/            business logic; interface repository didefinisikan di sini
  handler/            lapisan HTTP; interface service didefinisikan di sini
  middleware/         request id, access log, recover, secure headers, CORS, body limit, auth
  router/             definisi route dan rantai middleware
  security/           bcrypt dan penerbitan serta verifikasi JWT
  server/             http.Server dan graceful shutdown
  testutil/           fake, jam tetap, helper SQLite untuk test
test/
  integration/        repository, database, service dengan SQLite sungguhan
  e2e/                seluruh aplikasi lewat HTTP
```

Arah dependensi satu arah: `handler -> service -> repository -> model`. Interface didefinisikan oleh pemakainya (`service` mendefinisikan `UserRepository`, `handler` mendefinisikan `AuthService`), bukan oleh pembuatnya.

Handler tidak menulis response error sendiri. Handler cukup `return err`, dan `handler.NewErrorHandler` memetakan error ke status HTTP dan format JSON `{"error": {...}}` di satu tempat.

## Endpoint

| Method | Path | Auth | Hasil |
| --- | --- | --- | --- |
| GET | `/ping` | - | `{"status":"ok"}` |
| GET | `/swagger/` | - | Swagger UI; spesifikasi di `/swagger/openapi.json` |
| POST | `/api/v1/auth/signup` | - | `201` dengan pengguna baru |
| POST | `/api/v1/auth/login` | - | `200` dengan `token`, `token_type`, `expires_at`, `user` |
| POST | `/api/v1/auth/logout` | Bearer | `204` |
| GET | `/api/v1/auth/me` | Bearer | pengguna saat ini |

Route `/api/v1/snippets` tidak ada di service ini; route itu dimiliki service snippet.

## Konfigurasi

| Variabel | Default | Keterangan |
| --- | --- | --- |
| `JWT_SECRET` | wajib | minimal 32 karakter; harus sama dengan service snippet |
| `JWT_TTL` | `12h` | masa berlaku token |
| `HTTP_ADDR` | `:8081` | |
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
| Unit | di samping kode, `*_test.go` | tanpa tag | satu paket, tanpa I/O. Service memakai fake repository; handler memakai fake service. |
| Integration | `test/integration` | `-tags=integration` | GORM, migrasi, retry koneksi, dan service + repository + bcrypt + JWT pada SQLite in-memory. |
| E2E | `test/e2e` | `-tags=e2e` | `httptest.Server` dengan wiring yang sama dengan produksi, memakai klien HTTP sungguhan. |

Integration dan e2e memakai SQLite (`glebarez/sqlite`, tanpa cgo), bukan MySQL. Perilaku khas MySQL (collation, tipe kolom, pesan error driver) tidak tercakup oleh test ini.
