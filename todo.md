# TODO: Device Firmware Scan Service

## Requirements
- [ ] POST /v1/firmware-scans (Register scan, async analysis)
- [ ] PATCH /v1/findings/vulns (Add unique CVEs)
- [ ] GET /v1/findings/vulns (List unique CVEs)
- [ ] Distributed architecture support (replicas behind load balancer)
- [ ] Idempotent processing (handle duplicate requests)
- [ ] Scalability & High load handling

## Technology Stack
- **Language:** Golang
- **Database:** Postgres (SQLx, schema.sql)
- **Infrastructure:** Docker, Docker Compose
- **Framework:** Echo
- **Configuration:** ENV flags (including SIMULATE_FAILURE)

## Tasks
- [x] Initialize todo.md
- [ ] Initialize Go project (Standard structure)
- [ ] Setup Docker Compose (Postgres)
- [ ] Implement Firmware Registration API (Echo)
- [ ] Implement Asynchronous Worker with Postgres-based queue
- [ ] Implement CVE Registry API (PATCH/GET)
- [ ] Add simulation failure logic (ENV gated)
- [ ] Write Unit & Integration Tests (Testify + Testcontainers)
- [ ] Documentation (Architecture Notes)
