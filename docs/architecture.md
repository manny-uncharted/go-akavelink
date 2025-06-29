# go-akavelink Architecture

This document outlines the basic structure and design decisions for the project.

---

## 🎯 Purpose

`go-akavelink` is a lightweight HTTP API server written in Go, wrapping the Akave SDK directly to provide REST endpoints for file storage, retrieval, and management.

---

## 🧱 Project Structure

```
go-akavelink/
├── cmd/server/       # Entrypoint to the server (main.go)
├── internal/sdk/     # Akave SDK client wrapper logic
├── pkg/              # Shared public utilities (optional)
├── docs/             # Technical documentation and specs
```

---

## 🔄 Request Flow

```
Client --> go-akavelink HTTP API --> Akave SDK --> Akave Backend
```

---

## 🧩 Planned Modules

- Health check (`/health`)
- Bucket management:
  - `GET /buckets`
  - `GET /buckets/:id`
  - `POST /buckets/:id`
- File operations:
  - `GET /:bucket_id/files`
  - `GET /:bucket_id/files/:id`
  - `POST /:bucket_id/files/upload`
  - `GET /:bucket_id/files/:id/download`
- Auth and config layer
- Middleware (logging, CORS, etc.)

---

## 📌 Notes

- All SDK interactions will be wrapped in a thin abstraction (`internal/sdk/client.go`)
- The HTTP layer should remain stateless
- Follow Go idioms: small interfaces, dependency injection where needed, idiomatic error handling

---

Stay tuned for updated diagrams and flowcharts as the project evolves.
