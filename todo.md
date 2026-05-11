# TODO: Firmguard - Device Firmware Scan Service

## Requirements
- [ ] POST /v1/firmware-scans (Register scan, async analysis)
- [ ] PATCH /v1/findings/vulns (Add unique CVEs)
- [ ] GET /v1/findings/vulns (List unique CVEs)
- [ ] Distributed architecture support (replicas behind load balancer)
- [ ] Idempotent processing (handle duplicate requests)
- [ ] Scalability & High load handling

## Technology Stack
- **Language:** Golang
- **Database:** Postgres (SQLx, goose migrations)
- **Infrastructure:** Docker, Docker Compose
- **Framework:** Echo
- **Configuration:** ENV flags (including SIMULATE_FAILURE)
- **Quality:** 100% code coverage, go fmt, race detection

## Tasks
- [x] Initialize todo.md
- [x] Initialize Go project (Standard structure: cmd/api, internal/)
- [x] Setup Docker Compose & Goose migrations
- [x] Implement Firmware Registration API (POST /v1/firmware-scans)
- [x] Implement Asynchronous Worker Pool (Postgres-based queue)
- [x] Add simulation failure logic (ENV gated)
- [x] Achieve 100% code coverage for implemented logic
- [ ] Implement CVE Registry API (PATCH/GET)
- [ ] Final Review & Documentation

