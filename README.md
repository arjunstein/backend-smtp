# Centralized SMTP Email Sender Service (`backend-smtp`)

[![Go Version](https://img.shields.io/badge/go-1.25%2B-blue.svg)](https://golang.org)
[![Router](https://img.shields.io/badge/router-Chi%20v5-green.svg)](https://github.com/go-chi/chi)
[![ORM](https://img.shields.io/badge/orm-GORM-orange.svg)](https://gorm.io)

A production-grade centralized backend service for managing SMTP configurations and dispatching outgoing emails across multi-tenant applications and microservices. Acts as a centralized email dispatch agent designed with Clean Architecture principles.

---

## 🚀 Key Features

- **Clean Architecture & Separation of Concerns**: Strictly divided into Handler, Service, Repository, Model, and Package layers.
- **Go Chi Router v5**: Lightweight, composable, and 100% compliant with the standard Go `net/http` package.
- **SMTP Configuration Management (Single `smtp_configs` Table)**:
  - Register multiple SMTP accounts (Gmail, Mailtrap, Amazon SES, SendGrid, Postfix, Zimbra, etc.).
  - Automatic credential and password masking in all API responses.
  - Atomic transactional default toggling (`is_default = true`) ensuring only one active default configuration system-wide.
  - Active connection test endpoint (`POST /api/v1/smtp-configs/{id}/test`) for pinging and validating credentials without sending an email.
- **Centralized Email Dispatch Endpoint (`POST /api/v1/emails/send`)**:
  - Unified endpoint consumed by external microservices and client applications.
  - Automatic fallback to the active default SMTP configuration when `smtp_config_id` is omitted.
  - Support for both Plain Text and Rich HTML content.
  - Support for Base64-encoded file attachments.
- **Independent SMTP Engine**:
  - Native Go implementation supporting `STARTTLS` (port 587), `SSL/TLS` (port 465), and `Plain` (port 25).
  - RFC 5322 compliant MIME builder supporting `multipart/alternative` and `multipart/mixed`.
- **Cloud-Native Structured Observability (`log/slog`)**:
  - Real-time JSON structured logging to `stdout` (`event=email_dispatch`, `status=SUCCESS/FAILED`, duration, recipients, subject, error details).
  - Eliminates database table bloat and write lock contention.
- **Security & Middleware**:
  - Strict `X-API-Key` authentication.
  - Structured HTTP request logger (`[HTTP] status | latency | ip | method path`).
  - Cross-Origin Resource Sharing (CORS) middleware.
  - Panic recovery middleware.
- **Interactive Swagger Documentation**: Integrated at `/documentation/index.html`.

---

## 📂 Project Structure

```
backend-smtp/
├── cmd/
│   └── api/
│       └── main.go                 # Application bootstrap, Chi router, & graceful shutdown
├── internal/
│   ├── config/
│   │   └── config.go               # Environment configuration loader (.env) & MySQL DSN
│   ├── handler/
│   │   ├── health_handler.go       # Health check HTTP handler
│   │   ├── smtp_config_handler.go  # SMTP Config CRUD & test connection handler
│   │   └── email_handler.go        # Centralized email dispatch HTTP handler
│   ├── middleware/
│   │   ├── auth.go                 # X-API-Key authentication middleware
│   │   ├── cors.go                 # CORS configuration middleware
│   │   └── logger.go               # Structured HTTP request logging middleware
│   ├── model/
│   │   ├── smtp_config.go          # GORM SMTPConfig entity (smtp_configs table)
│   │   └── dto.go                  # Data Transfer Objects (DTOs) & input validation
│   ├── repository/
│   │   └── smtp_config_repo.go     # GORM database queries & transactions
│   └── service/
│       ├── smtp_config_service.go  # SMTP configuration business logic
│       ├── email_service.go        # Dispatch orchestration & structured stdout logging (slog)
│       └── mailer/
│           ├── mailer.go           # Native SMTP client engine (STARTTLS, SSL/TLS, Plain, MIME)
│           └── mailer_test.go      # MIME builder unit tests
├── pkg/
│   └── response/
│       ├── response.go             # Standardized API response helpers for Chi (net/http)
│       └── response_test.go        # Response helper unit tests
├── docs/                           # Swagger OpenAPI 2.0 specs (swagger.json, docs.go)
├── .env.example                    # Sample environment variables
├── .gitignore
├── AGENTS.md                       # AI coding agent & development guidelines
├── Makefile                        # Build, test, and development scripts
└── README.md
```

---

## 🛠️ Prerequisites

- **Go**: 1.25 or higher
- **MySQL**: 8.0 or higher

---

## ⚙️ Environment Configuration (`.env`)

Copy the sample environment file `.env.example` to `.env`:

```bash
cp .env.example .env
```

Update the variables to match your server and MySQL configuration:

```env
# Server Configuration
SERVER_PORT=8082
API_KEY=your-secret-api-key-here

# Database Configuration (MySQL)
DB_HOST=127.0.0.1
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_database_password
DB_NAME=smtp_service
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME_MINUTES=5
```

> **Note**: If `API_KEY` is left blank, API authentication is bypassed for local development convenience.

---

## 🚀 Running the Application

```bash
# 1. Run automated unit tests with race detection
make test

# 2. Build the binary executable
make build

# 3. Start the application
make run
# or run the compiled binary directly:
./bin/backend-smtp
```

The application automatically executes database schema migrations for `smtp_configs` on startup.

---

## 📖 Swagger API Documentation

Open the interactive Swagger UI in your browser:
```
http://localhost:8082/documentation/index.html
```

---

## 📡 API Endpoints Reference

All `/api/v1` endpoints (except public `/health`) require the `X-API-Key: <API_KEY>` HTTP header when configured.

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/health` / `/api/v1/health` | Service health and uptime check |
| `POST` | `/api/v1/smtp-configs` | Create a new SMTP configuration |
| `GET` | `/api/v1/smtp-configs` | List all SMTP configurations (passwords masked) |
| `GET` | `/api/v1/smtp-configs/{id}` | Get specific SMTP configuration by ID |
| `PUT` | `/api/v1/smtp-configs/{id}` | Update SMTP configuration fields |
| `DELETE` | `/api/v1/smtp-configs/{id}` | Delete an SMTP configuration |
| `PUT` | `/api/v1/smtp-configs/{id}/default` | Set an SMTP configuration as active system default |
| `POST` | `/api/v1/smtp-configs/{id}/test` | Test SMTP server connection & authentication (ping) |
| `POST` | `/api/v1/emails/send` | **Centralized email dispatch endpoint** |

---

## 💡 Usage Examples (cURL)

### 1. Create New SMTP Configuration
```bash
curl -X POST http://localhost:8082/api/v1/smtp-configs \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-secret-api-key-here" \
  -d '{
    "name": "Transactional Mailer",
    "host": "smtp.example.com",
    "port": 587,
    "username": "no-reply@example.com",
    "password": "your-app-password",
    "from_email": "no-reply@example.com",
    "from_name": "Notification System",
    "encryption": "STARTTLS",
    "is_active": true,
    "is_default": true
  }'
```

### 2. Test SMTP Connection
```bash
curl -X POST http://localhost:8082/api/v1/smtp-configs/1/test \
  -H "X-API-Key: your-secret-api-key-here"
```

### 3. Send Email (Centralized Consumption)
Omit `smtp_config_id` to automatically use the active default SMTP configuration:

```bash
curl -X POST http://localhost:8082/api/v1/emails/send \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-secret-api-key-here" \
  -d '{
    "to": ["user@example.com"],
    "cc": ["manager@example.com"],
    "subject": "Your Verification Code",
    "body": "<h1>Welcome!</h1><p>Your one-time verification code is <b>827103</b>.</p>",
    "is_html": true
  }'
```

To route through a specific SMTP configuration:
```bash
curl -X POST http://localhost:8082/api/v1/emails/send \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-secret-api-key-here" \
  -d '{
    "smtp_config_id": 2,
    "to": ["client@example.com"],
    "subject": "Monthly Statement",
    "body": "Please find your monthly statement attached.",
    "is_html": false
  }'
```

---

## 🧪 Testing & Makefile Commands

```bash
make test      # Run all unit tests with race detection
make build     # Compile the application binary to bin/backend-smtp
make run       # Run the application locally
make swagger   # Regenerate Swagger OpenAPI documentation
make tidy      # Clean up go.mod and go.sum dependencies
make clean     # Remove compiled binaries
```
