# Centralized SMTP Email Sender Service (`backend-smtp`)

[![Go Version](https://img.shields.io/badge/go-1.25%2B-blue.svg)](https://golang.org)
[![Router](https://img.shields.io/badge/router-Chi%20v5-green.svg)](https://github.com/go-chi/chi)
[![ORM](https://img.shields.io/badge/orm-GORM-orange.svg)](https://gorm.io)

Layanan backend terpusat untuk manajemen konfigurasi SMTP dan pengiriman email multi-tenant/multi-service. Bertindak sebagai centralized email sender agent dengan arsitektur bersih (Clean Architecture) mengikuti konvensi repositori `gymmaster-api`.

---

## 🚀 Fitur Utama

- **Clean Architecture & Separation of Concerns**: Terbagi jelas antara Handler, Service, Repository, Model, dan Package.
- **Go Chi Router v5**: Router ringan, composable, dan berbasis standard library `net/http`.
- **CRUD Konfigurasi SMTP**:
  - Pendaftaran multi-akun SMTP (Gmail, Mailtrap, SES, Sendgrid, Postfix, dll).
  - Masking otomatis kredensial/password pada respon API.
  - Pengaturan konfigurasi *Default* sistem dengan transaksi database atomik.
  - Endpoint uji koneksi/ping SMTP (`POST /api/v1/smtp-configs/{id}/test`).
- **Endpoint Konsumsi Pengiriman Terpusat (`POST /api/v1/emails/send`)**:
  - Konsumsi oleh berbagai microservice / aplikasi client lain.
  - Auto-fallback ke active default SMTP config jika ID konfigurasi tidak ditentukan.
  - Dukungan pesan Plain Text dan Rich HTML.
  - Dukungan lampiran dokumen / file (Attachment) berbasis Base64.
- **Engine Mailer Mandiri**:
  - Dukungan protokol `STARTTLS` (port 587), `SSL/TLS` (port 465), dan `Plain` (port 25).
  - Pembangunan pesan RFC 5322 MIME dengan `multipart/alternative` dan `multipart/mixed`.
- **Audit Logging**: Setiap pengiriman email tercatat otomatis ke tabel `email_logs` (status `SUCCESS`/`FAILED`, penerima, subjek, timestamp, dan detail error jika gagal).
- **Keamanan & Middleware**:
  - Otentikasi `X-API-Key`.
  - Request logging terstruktur (`[HTTP] status | latency | ip | method path`).
  - CORS Middleware.
  - Panic recoverer.
- **Dokumentasi Swagger OpenAPI**: Terintegrasi pada `/documentation/index.html`.

---

## 📂 Struktur Proyek

```
backend-smtp/
├── cmd/
│   └── api/
│       └── main.go                 # Inisialisasi aplikasi, Chi router, & graceful shutdown
├── internal/
│   ├── config/
│   │   └── config.go               # Pemanggilan environment variables & DSN MySQL
│   ├── handler/
│   │   ├── health_handler.go       # Health check handler
│   │   ├── smtp_config_handler.go  # Handler CRUD & test SMTP
│   │   └── email_handler.go        # Handler pengiriman email & log audit
│   ├── middleware/
│   │   ├── auth.go                 # X-API-Key auth middleware
│   │   ├── cors.go                 # CORS middleware
│   │   └── logger.go               # Structured HTTP logging middleware
│   ├── model/
│   │   ├── smtp_config.go          # GORM model konfigurasi SMTP
│   │   ├── email_log.go            # GORM model audit pengiriman email
│   │   └── dto.go                  # Data Transfer Objects & validasi payload
│   ├── repository/
│   │   ├── smtp_config_repo.go     # Query database SMTPConfig (GORM)
│   │   └── email_log_repo.go       # Query database EmailLog (GORM)
│   └── service/
│       ├── smtp_config_service.go  # Logika bisnis konfigurasi SMTP
│       ├── email_service.go        # Logika pengiriman & pencatatan log
│       └── mailer/
│           ├── mailer.go           # Engine SMTP client (STARTTLS, SSL/TLS, Plain, MIME)
│           └── mailer_test.go      # Unit test MIME builder
├── pkg/
│   └── response/
│       ├── response.go             # Helper format respon standar Chi (net/http)
│       └── response_test.go        # Unit test format respon
├── docs/                           # Swagger OpenAPI spec (swagger.json, docs.go)
├── .env.example                    # Template environment variables
├── .gitignore
├── AGENTS.md                       # Panduan arsitektur & konvensi agen
├── Makefile                        # Perintah build, test, & run
└── README.md
```

---

## 🛠️ Persyaratan Sistem

- **Go**: 1.25 atau lebih baru
- **MySQL**: 8.0 atau lebih baru

---

## ⚙️ Konfigurasi Environment (`.env`)

Salin file `.env.example` ke `.env`:

```bash
cp .env.example .env
```

Isi variabel sesuai konfigurasi server dan MySQL Anda:

```env
SERVER_PORT=8080
API_KEY=my-super-secret-key

DB_HOST=127.0.0.1
DB_PORT=3306
DB_USER=root
DB_PASSWORD=
DB_NAME=smtp_service
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME_MINUTES=5
```

---

## 🚀 Menjalankan Aplikasi

```bash
# 1. Jalankan unit test
make test

# 2. Kompilasi binary
make build

# 3. Jalankan server
make run
# atau langsung jalankan binary yang sudah di-compile:
./bin/backend-smtp
```

Aplikasi akan otomatis melakukan migrasi tabel (`smtp_configs`, `email_logs`) saat pertama kali dijalankan.

---

## 📖 Dokumentasi API (Swagger)

Akses dokumentasi interaktif Swagger melalui browser di:
```
http://localhost:8080/documentation/index.html
```

---

## 📡 Daftar Endpoint API

Semua endpoint `/api/v1` (kecuali `/health`) membutuhkan header `X-API-Key: <API_KEY>` jika diatur di `.env`.

| Method | Endpoint | Deskripsi |
|---|---|---|
| `GET` | `/health` / `/api/v1/health` | Health check layanan |
| `POST` | `/api/v1/smtp-configs` | Tambah konfigurasi SMTP baru |
| `GET` | `/api/v1/smtp-configs` | Ambil daftar semua konfigurasi SMTP |
| `GET` | `/api/v1/smtp-configs/{id}` | Ambil detail satu konfigurasi SMTP |
| `PUT` | `/api/v1/smtp-configs/{id}` | Update konfigurasi SMTP |
| `DELETE` | `/api/v1/smtp-configs/{id}` | Hapus konfigurasi SMTP |
| `PUT` | `/api/v1/smtp-configs/{id}/default` | Setel konfigurasi sebagai default sistem |
| `POST` | `/api/v1/smtp-configs/{id}/test` | Tes koneksi & otentikasi SMTP |
| `POST` | `/api/v1/emails/send` | **Endpoint Utama Pengiriman Email Terpusat** |
| `GET` | `/api/v1/emails/logs` | Ambil riwayat / log audit pengiriman email |
| `GET` | `/api/v1/emails/logs/{id}` | Ambil detail satu record log pengiriman |

---

## 💡 Contoh Penggunaan (cURL)

### 1. Menambahkan Konfigurasi SMTP
```bash
curl -X POST http://localhost:8080/api/v1/smtp-configs \
  -H "Content-Type: application/json" \
  -H "X-API-Key: my-super-secret-key" \
  -d '{
    "name": "Gmail Notifikasi",
    "host": "smtp.gmail.com",
    "port": 587,
    "username": "notifikasi@example.com",
    "password": "app-password-rahasia",
    "from_email": "notifikasi@example.com",
    "from_name": "Sistem Notifikasi",
    "encryption": "STARTTLS",
    "is_active": true,
    "is_default": true
  }'
```

### 2. Menguji Koneksi SMTP
```bash
curl -X POST http://localhost:8080/api/v1/smtp-configs/1/test \
  -H "X-API-Key: my-super-secret-key"
```

### 3. Mengirimkan Email (Dikonsumsi oleh Sistem Lain)
Endpoint ini dapat dipanggil tanpa `smtp_config_id` untuk otomatis memakai SMTP default:

```bash
curl -X POST http://localhost:8080/api/v1/emails/send \
  -H "Content-Type: application/json" \
  -H "X-API-Key: my-super-secret-key" \
  -d '{
    "to": ["user@example.com"],
    "cc": ["supervisor@example.com"],
    "subject": "Kode Verifikasi Akun Anda",
    "body": "<h1>Halo!</h1><p>Kode verifikasi Anda adalah <b>827103</b>.</p>",
    "is_html": true
  }'
```

Jika ingin menggunakan konfigurasi SMTP tertentu:
```bash
curl -X POST http://localhost:8080/api/v1/emails/send \
  -H "Content-Type: application/json" \
  -H "X-API-Key: my-super-secret-key" \
  -d '{
    "smtp_config_id": 2,
    "to": ["client@example.com"],
    "subject": "Faktur Tagihan",
    "body": "Terlampir faktur tagihan bulan ini.",
    "is_html": false
  }'
```

### 4. Melihat Riwayat Log Pengiriman
```bash
curl -X GET "http://localhost:8080/api/v1/emails/logs?limit=10&offset=0" \
  -H "X-API-Key: my-super-secret-key"
```

---

## 🧪 Pengujian Unit

Jalankan seluruh test suite dengan race detector:
```bash
make test
```
