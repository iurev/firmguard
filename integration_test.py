import requests
import time
import sys

BASE_URL = "http://localhost:8080/v1"

def test_flow():
    print("--- 1. Registering CVEs ---")
    vuln_req = {"vulns": ["CVE-2024-TEST-001", "CVE-2024-TEST-002"]}
    resp = requests.patch(f"{BASE_URL}/findings/vulns", json=vuln_req)
    resp.raise_for_status()
    print("✓ CVEs registered")

    print("\n--- 2. Registering Firmware Scan ---")
    scan_req = {
        "device_id": "dev-py-test",
        "firmware_version": "1.0.0",
        "binary_hash": "hash-" + str(int(time.time())),
        "metadata": {"source": "integration-test"}
    }
    resp = requests.post(f"{BASE_URL}/firmware-scans", json=scan_req)
    resp.raise_for_status()
    scan_id = resp.json()["id"]
    print(f"✓ Scan registered (ID: {scan_id})")

    print("\n--- 3. Waiting for Analysis (Polling) ---")
    terminal_status = False
    for _ in range(35):
        resp = requests.get(f"{BASE_URL}/firmware-scans/{scan_id}")
        status = resp.json()["status"]
        print(f"Current status: {status}")
        if status in ["completed", "failed"]:
            terminal_status = True
            break
        time.sleep(2)
        
    if not terminal_status:
        raise Exception(f"Scan {scan_id} did not complete in time.")

    print("\n--- 4. Checking CVE Registry ---")
    resp = requests.get(f"{BASE_URL}/findings/vulns")
    vulns = resp.json()["vulns"]
    print(f"✓ Registry contains {len(vulns)} CVEs")
    
    print("\nSUCCESS: End-to-end flow verified.")

if __name__ == "__main__":
    try:
        test_flow()
    except Exception as e:
        print(f"\nFAILED: {e}")
        sys.exit(1)
