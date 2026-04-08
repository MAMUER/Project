#!/usr/bin/env python3
"""
Fitness Platform — Load Test (cross-platform)

Requires: k6 installed (https://k6.io/docs/getting-started/installation/)

Usage:
    python scripts/load-test.py
    python scripts/load-test.py --duration 2m --vus 50
    python scripts/load-test.py --base-url https://localhost:8443
"""

import os
import sys
import shutil
import argparse
import subprocess
from pathlib import Path

GREEN = "\033[92m"
YELLOW = "\033[93m"
RED = "\033[91m"
CYAN = "\033[96m"
RESET = "\033[0m"
BOLD = "\033[1m"

SCRIPT_DIR = Path(__file__).parent
LOAD_TEST_JS = SCRIPT_DIR / "load-test" / "load-test.js"


def main():
    parser = argparse.ArgumentParser(description="Fitness Platform Load Test (k6 wrapper)")
    parser.add_argument("--base-url", default="https://localhost:8443", help="API base URL")
    parser.add_argument("--duration", default="2m", help="Test duration (e.g. 1m, 5m, 30s)")
    parser.add_argument("--vus", type=int, default=50, help="Number of virtual users")
    args = parser.parse_args()

    # Check k6
    k6_path = shutil.which("k6")
    if not k6_path:
        print(f"{RED}k6 not found. Install it: https://k6.io/docs/getting-started/installation/{RESET}")
        sys.exit(1)

    if not LOAD_TEST_JS.exists():
        print(f"{RED}Load test script not found: {LOAD_TEST_JS}{RESET}")
        sys.exit(1)

    print(f"\n{BOLD}{CYAN}{'=' * 50}{RESET}")
    print(f"{BOLD}{CYAN}   FITNESS PLATFORM — LOAD TEST{RESET}")
    print(f"{BOLD}{CYAN}{'=' * 50}{RESET}")
    print(f"  k6        : {k6_path}")
    print(f"  Script    : {LOAD_TEST_JS}")
    print(f"  Base URL  : {args.base_url}")
    print(f"  Duration  : {args.duration}")
    print(f"  VUs       : {args.vus}")
    print()

    cmd = [
        k6_path, "run",
        "--out", "json=results.json",
        str(LOAD_TEST_JS),
        "--env", f"BASE_URL={args.base_url}",
        "--env", f"DURATION={args.duration}",
        "--env", f"VUS={args.vus}",
        "--no-usage-report",
    ]

    print(f"{YELLOW}Running: {' '.join(cmd[:6])}...{RESET}\n")

    try:
        result = subprocess.run(cmd, timeout=600)
        if result.returncode == 0:
            print(f"\n{GREEN}{BOLD}  LOAD TEST COMPLETED!{RESET}\n")
        else:
            print(f"\n{RED}{BOLD}  LOAD TEST FAILED (exit code {result.returncode}){RESET}\n")
            sys.exit(result.returncode)
    except subprocess.TimeoutExpired:
        print(f"\n{RED}Load test timed out (10 min){RESET}")
        sys.exit(1)
    except KeyboardInterrupt:
        print(f"\n{YELLOW}Load test interrupted by user{RESET}")
        sys.exit(130)


if __name__ == "__main__":
    main()
