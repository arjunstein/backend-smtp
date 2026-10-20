# AGENTS.md — Development Guidelines & Best Practices

This document provides system guidelines, architectural rules, and operational best practices for AI coding agents working on the **Centralized SMTP Email Sender Service (`backend-smtp`)** project.

---

## 1. Project Overview & Architecture

- **Project**: Centralized SMTP Email Sender Service
- **Repository**: `github.com/arjunstein/backend-smtp`
- **Language**: Go 1.25+
- **Router / Web Framework**: Go Chi (`github.com/go-chi/chi/v5`)
- **Database & ORM**: MySQL 8.0+ via GORM (`gorm.io/gorm`, `gorm.io/driver/mysql`)
- **Documentation**: Swagger / OpenAPI 2.0 (`swaggo/swag`, `swaggo/http-swagger/v2`)

### Layered Clean Architecture
Code is strictly organized by concerns. Do not bypass layers:
```
cmd/api/main.go            # Dependency injection, Chi router, server lifecycle & graceful shutdown
├── internal/
│   ├── config/            # Environment variable loading & validation (.env, DSN, JWT)
│   ├── handler/           # HTTP transport, JSON unmarshaling, Swagger annotations (auth, health, smtp, email)
│   ├── middleware/        # Request logging, CORS, API Key and JWT authentication
│   ├── service/           # Business logic, validation, fallback orchestration, auth, and mailer interface
│   │   └── mailer/        # RFC 5322 MIME builder, SMTP client (STARTTLS, SSL/TLS, Plain)
│   ├── repository/        # GORM database queries & transactions (smtp_configs, users)
│   └── model/             # GORM entities (User, SMTPConfig), DTOs, request/response models
└── pkg/
    └── response/          # Standardized API response format helpers for Chi (net/http)
```

---

## 2. Core Business Rules & Email Delivery

### SMTP Configuration Management
- **Security & Masking**: Passwords are saved directly to MySQL, but MUST NEVER be exposed in JSON GET/list responses.
- **Uniqueness**: Config `name` is unique (`uniqueIndex`).
- **Default Config**: Only one configuration can be active default (`is_default = true`). When `SetDefault` is invoked, other configs must be set to `is_default = false` within an atomic database transaction.
- **Connection Test**: The `POST /api/v1/smtp-configs/{id}/test` endpoint executes an active ping (`Noop()`) and authentication check without sending an email.

### Centralized Email Dispatch Endpoint
- **Endpoint**: `POST /api/v1/emails/send`
- **Routing**:
  1. If `smtp_config_id` is supplied in the request, the agent resolves that specific configuration (rejecting if inactive).
  2. If omitted, the agent resolves the system default active configuration.
  3. If no active configuration exists, the agent returns a clear error.
- **Structured Observability**: Every send attempt is emitted to `stdout` via standard Go `log/slog` in JSON format (`event=email_dispatch`, `status=SUCCESS/FAILED`, duration, recipient, subject, error details), eliminating table bloat and write lock contention in MySQL.

---

## 3. Swagger & Documentation Standards

- Interactive documentation is served at `/documentation/index.html` (with redirects from `/documentation` and `/swagger/*`).
- Always regenerate Swagger docs after modifying handlers, DTOs, or Swagger annotations:
  ```bash
  make swagger
  ```
- Use generic dummy data (`no-reply@example.com`, `smtp.gmail.com`, `user@example.com`) in all examples. Never use production credentials.

---

## 4. Security & Safety Guidelines

1. **Authentication**:
   - All endpoints under `/api/v1` (except public `/health`) require the `X-API-Key` HTTP header when `API_KEY` is configured in `.env`.
2. **Credentials Protection**:
   - NEVER commit `.env`, credentials, or private keys to Git.
   - Always sanitize API responses to hide SMTP passwords.
3. **Command Safety**:
   - Never execute destructive commands (`rm -rf`, `git reset --hard`, `git push --force`, `drop table/database`) without explicit user confirmation.

---

## 5. Development & Testing Workflow

```bash
# Run unit and race-detector tests
make test

# Build executable
make build

# Run application
make run

# Clean dependencies
make tidy
```
