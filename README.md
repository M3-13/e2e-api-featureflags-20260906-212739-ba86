# e2e-api-featureflags

A feature-flag service implemented as a REST API in Go. Flags are managed in a
thread-safe in-memory store and can be evaluated per user via a deterministic
rollout decision. The service uses only the Go standard library.

## Tech Stack

- **Language**: Go (>= 1.22)
- **Framework**: `net/http` (standard library, Go 1.22 `ServeMux` method/pattern routing)
- **Storage**: in-memory with `sync.RWMutex`
- **Testing**: `net/http/httptest`

## Install

No external dependencies. Requires Go 1.22 or newer.

```sh
go version
```

## Run

Start the service (it binds to the port given by `PORT`, default `8080`):

```sh
go run .
```

or, to choose a port explicitly:

```sh
PORT=9000 go run .
```

## Endpoints

All endpoints respond with `application/json`.

| Method | Path                       | Description                                              |
| ------ | -------------------------- | -------------------------------------------------------- |
| POST   | `/flags`                   | Create a flag (`201`) or reject invalid/duplicate input |
| GET    | `/flags`                   | List all flags                                           |
| GET    | `/flags/{key}`             | Get a single flag                                        |
| PUT    | `/flags/{key}`             | Update a flag                                            |
| DELETE | `/flags/{key}`             | Delete a flag                                            |
| GET    | `/flags/{key}/evaluate`    | Evaluate a flag for a user (`?user={id}`)                |
| GET    | `/healthz`                 | Health check — responds `200 {"status":"ok"}`            |

## Configuration

| Variable | Default | Description                          |
| -------- | ------- | ------------------------------------ |
| `PORT`   | `8080`  | Port the HTTP server binds to        |

## Features

- Thread-safe in-memory flag store
- Deterministic per-user rollout evaluation
- Input validation with structured `{"error": ...}` responses
- Access logging middleware (method, path, status code)
- Health check endpoint

## Versioning

The service follows **Semantic Versioning (SemVer)**: `MAJOR.MINOR.PATCH`. The
current version is reported by the health endpoint (`GET /healthz`). Security
fixes are shipped as new releases; see [SECURITY.md](./SECURITY.md) for the
update/patch procedure and the support window.

## Security & Privacy

Security and privacy are documented in dedicated files:

- **[SECURITY.md](./SECURITY.md)** — security contact, vulnerability reporting
  procedure, support window, update/patch procedure, and binding operating
  requirements (TLS-terminating reverse proxy required, no security decisions).
- **[PRIVACY.md](./PRIVACY.md)** — description of how the `user` parameter is
  processed (purpose, data category, legal basis, retention, recipients,
  data-subject rights).

A minimal SPDX software bill of materials is available in
**[sbom.json](./sbom.json)** (no external dependencies).
