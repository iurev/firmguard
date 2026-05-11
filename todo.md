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
- [x] Database: Postgres (SQLx, goose migrations)
- [x] Infrastructure: Docker, Docker Compose
- [x] Framework: Echo
- [x] Configuration: ENV flags (including SIMULATE_FAILURE)
- [x] Quality: 100% code coverage, go fmt, race detection

## Tasks
- [x] Initialize todo.md
- [x] Initialize Go project (Standard structure: cmd/api, internal/)
- [x] Setup Docker Compose & Goose migrations
- [x] Implement Firmware Registration API (POST /v1/firmware-scans)
- [x] Implement Asynchronous Worker Pool (Postgres-based queue)
- [x] Add simulation failure logic (ENV gated)
- [x] Achieve 100% code coverage for implemented logic
- [x] Implement CVE Registry API (PATCH/GET)
- [x] Final Review & Documentation
- [ ] Integrate River for background jobs (Retries/Persistence)
- [ ] 100% code coverage for River worker and integration
