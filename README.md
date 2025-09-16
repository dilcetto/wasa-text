# WASAText

WASAText is a lightweight messaging platform built for the Web and Software Architecture course. It pairs a Go API backed by SQLite with a Vue 3 single page application and demonstrates authentication, conversation management, and messaging features suitable for the exam project we just completed. 

## Features
- Username-based onboarding with self-registration and stateless JWT authentication.
- SQLite-backed persistence for users, direct and group conversations, and message receipts.
- Direct chats and group conversations with photo, rename, add/invite, and leave operations.
- Rich messaging with text or photo attachments, delivery/read receipts, deletion, and forwarding.
- Emoji reactions aggregated per message.
- User and conversation search plus profile updates (username and avatar upload).
- Vue 3 SPA consuming the REST API defined in `doc/api.yaml`.

## Tech Stack
- Go 1.17 with `httprouter`, `logrus`, and `sqlite3`.
- SQLite database auto-initialised by `service/database`.
- Vue 3 + Vite, packaged with Yarn 4 using the offline mirror stored in `.yarn`.
- Docker-based Node 20 environment via `open-node.sh` for frontend development.
- Dockerfiles for backend and frontend deployment examples.

## Technologies Used

### Backend
- **Go 1.17** – API server, business logic
- **SQLite** – Embedded database
- **httprouter** – Lightweight router
- **logrus** – Structured logging
- **uuid** – Unique identifier generation

### Frontend
- **Vue 3 + Vite** – SPA UI framework and bundler
- **Vue Router** – Page navigation
- **Axios** – HTTP client
- **Bootstrap + custom CSS** – Responsive design

## Repository Layout
- `cmd/webapi/` – entrypoint that wires configuration, logging, database, and the HTTP server.
- `service/api/` – HTTP handlers (auth, profile, conversations, messages, reactions, groups).
- `service/database/` – SQLite persistence layer and schema bootstrap.
- `service/components/` – shared request/response schemas.
- `webui/` – Vue SPA, components, router, Axios client, and build tooling.
- `doc/api.yaml` – full OpenAPI 3 specification of the REST endpoints.
- `demo/config.yml` – sample configuration values for the API server.
- `Dockerfile.backend` / `Dockerfile.frontend` – container images for deployment.

## Getting Started

### Prerequisites
- Go 1.17 or newer.
- Docker (recommended for `open-node.sh`) or a local Node 20 + Yarn 4.5 installation.
- SQLite is bundled via CGO, so no extra service needs to be running.

### Backend
- Start the API server with `go run ./cmd/webapi`. By default it listens on `http://localhost:3000` and stores data in `/tmp/decaf.db`.
- Override settings via CLI flags or environment variables as defined in `cmd/webapi/load-configuration.go`. Example: `CFG_DB_FILENAME=./wasa.db go run ./cmd/webapi --cfg.web.apihost=127.0.0.1:3000`.
- The first run automatically bootstraps the database schema. Logs and graceful shutdown handling are managed for you.

Try a quick smoke test:

```bash
curl -X POST http://localhost:3000/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice"}'
```

The response returns the created user and a bearer token.

### Frontend
Run `./open-node.sh` to enter an ephemeral Node 20 container already mounted at `/src/webui`.

Inside the container start the dev server with `yarn run dev` and open `http://localhost:5173`. The command installs dependencies on demand thanks to the Yarn offline mirror.

For a production bundle run `yarn run build-prod`. To embed the SPA inside the Go binary run `yarn run build-embed` before building the backend with:

```bash
go build -tags webui ./cmd/webapi/
```

## Docker Images
- `Dockerfile.backend` builds the API service (optionally after embedding the web UI).
- `Dockerfile.frontend` serves the SPA standalone if you prefer to host it separately.

## API Documentation
The REST contract lives in `doc/api.yaml`. Preview it with any OpenAPI viewer, e.g.:

```bash
npx @redocly/openapi-cli preview-docs doc/api.yaml
```

Or point Swagger UI to the file.

Endpoints cover login, profile management, conversation discovery, message operations (send, forward, delete, set status), reactions, and group management.

## Testing
Run `go test ./...` to execute the backend unit tests and ensure the project still builds.

## Production Notes
- Replace the development secret in `service/api/token.go` (`jwtKey`) before deploying a public instance.
- Update `webui/vite.config.js` if the API is exposed on a URL other than `http://localhost:3000`; the `__API_URL__` constant controls the Axios base URL.
- Consider serving the API behind TLS and configuring reverse proxies/CORS as needed for your hosting environment.
- Remember to persist the SQLite database file or move to an external database if you expect multiple instances.

## License
This project is licensed under the MIT License – see the [LICENSE](LICENSE) file for details.

## Maintainers

This project was built by Dana Rabandiyar as part of the Web and Software Architecture course at Sapienza University of Rome (2025). For questions or reuse, feel free to fork or reference the codebase.
