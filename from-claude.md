I'll map each requirement to the exact text from the original document, then visualize the data flows.Let me lay out every requirement traced back to the original text, then show the architecture.

## Functional requirements

**FR1 — Accept and validate scan registrations**
> "Accept and validate incoming scan registrations."
> "POST /v1/firmware-scans" — payload has `device_id`, `firmware_version`, `binary_hash`, `metadata`.

**FR2 — Idempotency (no duplicate processing for same firmware)**
> "Ensure that duplicate scan requests do not result in redundant processing."
> "Devices may retry requests due to unreliable network conditions."
> "Devices may send repeated requests for the same firmware."

→ Solved with `UNIQUE (device_id, binary_hash)` + `INSERT ... ON CONFLICT DO NOTHING` on the `scans` table.

**FR3 — Trigger async analysis after registration**
> "Trigger an asynchronous analysis process after a scan is registered."
> "The analysis process can be simulated."
> "Once a scan request is registered, the firmware must be analyzed asynchronously."

→ Solved by inserting a row into `jobs` in the same transaction as the `scans` insert. Worker pool pulls from it.

**FR4 — Persist scan/analysis state so device only updates after success**
> "Persist relevant data so that the firmware state of the device can be updated only after analysis has completed successfully."
> "Update the device's security status only after the analysis has completed successfully."

→ `devices` table holds the *current validated* firmware. It only gets updated when a job reaches `validated` status. `scans` table tracks the in-flight scan separately.

**FR5 — CVE registry append + dedupe**
> "PATCH /v1/findings/vulns — Append new IDs to the global registry. Duplicates must be removed — the final registry must contain only unique IDs."

→ `vulns` table with `cve_id` as PRIMARY KEY. Bulk `INSERT ... ON CONFLICT (cve_id) DO NOTHING`.

**FR6 — CVE registry read**
> "GET /v1/findings/vulns — Return the current list of all unique CVE IDs in the system."

→ Simple `SELECT cve_id FROM vulns ORDER BY cve_id`.

## Non-functional / distributed requirements

**NFR1 — Burst handling**
> "The platform may experience bursts where thousands of devices report their firmware within a short period of time."

→ Async job processing decouples request acceptance from analysis. Postgres connection pool absorbs writes. Workers drain at their own pace.

**NFR2 — Long-running analysis**
> "Firmware analysis may take several seconds to complete."
> "long-running analysis processes."

→ Analysis runs on workers, not on the request thread. API returns `202 Accepted` immediately.

**NFR3 — Tolerate temporary failures (implies retries)**
> "The system should tolerate temporary failures in dependent components."

→ Job retry with `attempts`/`max_attempts`, exponential backoff via `next_run_at`, lease-based recovery (`running_until`) for crashed workers.

**NFR4 — Multiple replicas behind a load balancer**
> "Scaling: Must work correctly when running multiple replicas behind a load balancer."
> "Synchronization: Use any synchronization mechanism, architectural patterns, and/or infrastructure you consider appropriate to ensure all replicas share the same state."

→ API replicas are stateless. All state in Postgres — Postgres IS the synchronization mechanism.

**NFR5 — No race conditions / no duplicate entries across replicas**
> "Concurrency: Concurrent requests to different replicas must not result in race conditions or duplicate entries."

→ `ON CONFLICT DO NOTHING` on unique constraints (atomic), `FOR UPDATE SKIP LOCKED` on jobs (no worker races).

## Implicit requirements you'd be smart to address

**IR1 — Scan-to-CVE linkage in the analysis flow**
> "As the platform scans devices, various analysis engines identify known vulnerabilities."

The text implies analysis produces CVEs. Not explicit, but the system should reflect this. → `findings` JSONB column on the `scans` row, populated by the worker.

**IR2 — Network instability between device and API**
> "unreliable network conditions"
> "Devices may retry requests due to unreliable network conditions."

→ Covered by FR2's idempotency story.

**IR3 — Async response shape**
The "only update after analysis completes" requirement combined with "may take several seconds" implies POST must return immediately. → `202 Accepted` with a scan ID. GET endpoint returns the previous validated firmware until the new one is verified.

## Data flow diagram## What flows where

**Scan registration path** (covers FR1, FR2, FR3, FR4, NFR2, NFR4):
1. Device POSTs scan → load balancer picks any API replica
2. API replica validates, then in **one transaction**: `INSERT scans ON CONFLICT DO NOTHING` + `INSERT jobs (scan_id, status='pending')`
3. API returns `202 Accepted` with the scan ID immediately

**Async analysis path** (covers FR3, FR4, NFR3):
1. Worker runs: `SELECT ... FROM jobs WHERE status='pending' ... FOR UPDATE SKIP LOCKED LIMIT 1`
2. Marks job `running`, sets `running_until = now() + lease`
3. Runs simulated analysis → produces some CVE IDs
4. On success, **one transaction**: update `scans.findings`, `INSERT vulns ON CONFLICT DO NOTHING`, update `devices` row to validated firmware, mark job `done`
5. On failure: increment `attempts`, set `next_run_at = now() + backoff`, status back to `pending` (or `failed` if max hit)

**CVE registry path** (covers FR5, FR6, NFR5):
1. External analyzer PATCHes a batch of CVEs → API replica → one batch `INSERT INTO vulns ... ON CONFLICT DO NOTHING`
2. Dashboard GETs the registry → `SELECT cve_id FROM vulns`

## Where each requirement is enforced in the diagram

| Component | Requirements it handles |
|---|---|
| Load balancer + stateless replicas | NFR1, NFR4 |
| `scans.UNIQUE(device_id, binary_hash)` | FR2, IR2 |
| `jobs` table + SKIP LOCKED | FR3, NFR1, NFR4, NFR5 |
| Job retry fields (`attempts`, `next_run_at`, `running_until`) | NFR3 |
| Separation of `scans` and `devices` | FR4, IR3 |
| `vulns.cve_id PK` + `ON CONFLICT DO NOTHING` | FR5, NFR5 |
| `scans.findings` JSONB | IR1 |
| Worker pool consuming from Postgres | FR3, NFR1, NFR2 |

Everything from the original spec lands somewhere concrete in this design. The only thing I'd add to your architecture notes is one sentence on what's deliberately *out of scope* — e.g. "No `scan_findings` join table, since the prompt's CVE API is flat" — so the reviewer sees you decided rather than forgot.

Want me to draft the actual SQL schema next, or the docker-compose layout?