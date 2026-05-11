# Backend Engineering Challenge

## Device Firmware Scan Registration

### Context

Our platform monitors firmware integrity across a large fleet of industrial and embedded devices.
Devices periodically provide information about the firmware currently running on the device so that it can be analyzed and validated against supply chain records.

When a device reports its firmware information, the platform must:

1. Register the device firmware.
2. Trigger an analysis process that validates the firmware.
3. Update the device's security status only after the analysis has completed successfully.

In production, the platform must handle a high volume of devices, unreliable network conditions, and long-running analysis processes.

### The Problem

You are asked to implement a service responsible for registering firmware scan requests coming from devices.

Devices send information describing the firmware currently running on the device, along with metadata about the system.

Once a scan request is registered, the firmware must be analyzed asynchronously.
Only after the analysis is completed successfully should the device record reflect the validated firmware state.

### APIs

#### 1. Firmware Scan Registration

`POST /v1/firmware-scans`

Example payload:

```json
{
  "device_id": "device-123",
  "firmware_version": "2.4.1",
  "binary_hash": "ab34c9f...",
  "metadata": {
    "components": [...],
    "hardware_model": "X1000",
    "additional_info": {...}
  }
}
```

**Notes:**
- The metadata field may contain a large JSON payload describing device components and hardware characteristics.
- Devices may retry requests due to unreliable network conditions.

---

#### 2. Distributed CVE Registry

As the platform scans devices, various analysis engines identify known vulnerabilities. To provide a unified security posture, the system must maintain a centralized, real-time registry of all unique CVE IDs discovered across the entire fleet. This registry serves as a "Source of Truth" for security dashboards and reporting tools.

If you are curious about what vulnerabilities are, visit the [MITRE CVE Database](https://cve.mitre.org/).

Implement the following APIs:

**`PATCH /v1/findings/vulns`**

Append new IDs to the global registry. Duplicates must be removed — the final registry must contain only unique IDs.

Example payload:

```json
{ "vulns": ["CVE-001", "CVE-002", ...] }
```

Example — after these two requests:

```sh
curl -L -X PATCH 'http://backend/v1/findings/vulns' \
  -H 'Content-Type: application/json' \
  -d '{"vulns":["CVE-001","CVE-002"]}'

curl -L -X PATCH 'http://backend/v1/findings/vulns' \
  -H 'Content-Type: application/json' \
  -d '{"vulns":["CVE-002","CVE-003"]}'
```

A subsequent GET:

```sh
curl -L 'http://backend/v1/findings/vulns'
```

returns:

```json
{
  "vulns": ["CVE-001", "CVE-002", "CVE-003"]
}
```

**`GET /v1/findings/vulns`**

Return the current list of all unique CVE IDs in the system.

**Distributed Architecture Requirements:**
- **Scaling:** Must work correctly when running multiple replicas behind a load balancer.
- **Synchronization:** Use any synchronization mechanism, architectural patterns, and/or infrastructure you consider appropriate to ensure all replicas share the same state.
- **Concurrency:** Concurrent requests to different replicas must not result in race conditions or duplicate entries.

## Functional Requirements

Your service should:

- Accept and validate incoming scan registrations.
- Ensure that duplicate scan requests do not result in redundant processing.
- Trigger an asynchronous analysis process after a scan is registered.
- The analysis process can be simulated.
- Persist relevant data so that the firmware state of the device can be updated only after analysis has completed successfully.

## Non-Functional Considerations

Your implementation should consider the following conditions:

- Devices may send repeated requests for the same firmware.
- Multiple requests for the same device may arrive close together.
- Firmware analysis may take several seconds to complete.
- The system should tolerate temporary failures in dependent components.
- The platform may experience bursts where thousands of devices report their firmware within a short period of time.

## Technical Scope

We suggest using Go, Node (TS), but you may choose any programming language and supporting technologies you prefer.

Your solution should run locally and include clear instructions on how to start the service.

### Deliverables

Please provide:

- Source Code
- A working implementation of the service.
- Setup Instructions
- Instructions for running the project locally.
- Architecture Notes:
  Include a short document explaining:
    - The main architectural decisions you made
    - How your design handles duplicate or repeated requests
    - How asynchronous processing is coordinated
    - How the system would behave under high load
    - What changes would be needed to support significantly more devices

### Expectations

The goal of this exercise is not to build a complete production system, but to demonstrate how you approach:

- backend system design
- concurrency and data consistency
- asynchronous workflows
- reliability and scalability concerns

Keep the implementation reasonably simple while focusing on clarity and correctness.

## Submission

Upload the complete project to a public or private repository that can be accessed for review.

Include a document with clear instructions for evaluating the solution.

Good luck with the challenge! 🚀

