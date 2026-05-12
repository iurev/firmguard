# FirmGuard Service

Distributed firmware integrity monitoring and CVE registry service.

For a detailed technical overview, see [ARCHITECTURE.md](ARCHITECTURE.md).

## Operational Guide: Local Setup & Testing

### 1. Prerequisites
- **Go 1.25+**
- **Docker** (for Postgres)
- **Goose** (for migrations)

### 2. Start Infrastructure
```bash
make docker-run
```
- Starts Postgres via `docker-compose`.
- Runs `migrate-up` to create tables (`firmware_scans`, `vulnerabilities`, `river_jobs`).

### 3. Run Application
```bash
make run
```
- Starts the Echo API and River worker.
- API: `http://localhost:8080`.

### 4. Verify Implementation

**Update CVE Registry (Prefill Data)**
```bash
curl -X PATCH http://localhost:8080/v1/findings/vulns \
  -H "Content-Type: application/json" \
  -d '{"vulns":["CVE-2024-0001", "CVE-2024-0002"]}'
```
- *Note:* Prefilling the registry allows the simulated analysis to "discover" known vulnerabilities.

**Register a Scan (Asynchronous)**
```bash
curl -X POST http://localhost:8080/v1/firmware-scans \
  -H "Content-Type: application/json" \
  -d '{"device_id":"dev-123","firmware_version":"1.0.0","binary_hash":"abc...","metadata":{"model":"X1"}}'
```
- Returns **`202 Accepted`** for new registrations.
- Returns **`200 OK`** for duplicates (idempotent).

**Check Scan Status**
```bash
curl http://localhost:8080/v1/firmware-scans/1
```
- Transitions from `pending` → `completed` or `failed` after simulation (1-59s).

### 5. Automated Tests
```bash
# Full test suite (includes DB integration tests)
make test

# Integration Script (requires Python)
uv run --with requests integration_test.py
```

## Makefile Commands

- `make build`: Build the application.
- `make run`: Run the application locally.
- `make docker-run`: Start DB container and run migrations.
- `make docker-down`: Shutdown DB container.
- `make test`: Run Go test suite.
- `make clean`: Remove build artifacts.
