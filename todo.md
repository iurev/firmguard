# TODO: Firmguard - Device Firmware Scan Service

## Requirements
- [x] POST /v1/firmware-scans (Register scan, async analysis)
- [x] PATCH /v1/findings/vulns (Add unique CVEs)
- [x] GET /v1/findings/vulns (List unique CVEs)
- [x] Distributed architecture support (replicas behind load balancer)
- [x] Idempotent processing (handle duplicate requests)
- [x] Scalability & High load handling

## Technology Stack
- [x] Language: Golang
- [x] Database: Postgres (pgx, goose migrations)
- [x] Infrastructure: Docker, Docker Compose
- [x] Framework: Echo
- [x] Configuration: ENV flags (including SIMULATE_FAILURE)
- [x] Quality: 100% code coverage, go fmt, race detection

## Tasks
- [x] Initialize todo.md
- [x] Setup Docker Compose & Goose migrations
- [x] Implement Firmware Registration API (POST /v1/firmware-scans)
- [x] Implement Asynchronous Worker Pool (River/Postgres)
- [x] Implement CVE Registry API (PATCH/GET)
- [x] Add simulation failure logic (30/30/40 outcomes)
- [x] Align with `from-claude.md` (Atomic ON CONFLICT, 202 Accepted)
- [x] Refactor to use `pgx` consistently (remove `sqlx`)
- [x] Achieve 100% code coverage for api, service, worker, repository
- [x] Maintain 100% coverage after refactoring
- [x] Final Review & Documentation
