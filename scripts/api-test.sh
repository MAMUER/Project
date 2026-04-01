#!/bin/bash
# scripts/load-test.sh
# Load Test Runner for Linux

set -e

BASE_URL="${1:-http://localhost:8080}"
DURATION="${2:-2m}"
VUS="${3:-50}"

# Help
if [ "$1" == "-h" ] || [ "$1" == "--help" ]; then
    cat << EOF
========================================
   FITNESS PLATFORM - LOAD TEST
========================================

Usage: ./scripts/load-test.sh [options]

Options:
  BASE_URL   API base URL (default: http://localhost:8080)
  DURATION   Test duration (default: 2m)
  VUS        Virtual users (default: 50)

Examples:
  ./scripts/load-test.sh
  ./scripts/load-test.sh http://localhost:8080 5m 100
  ./scripts/load-test.sh http://api.example.com 3m 75

========================================
EOF
    exit 0
fi

echo "========================================"
echo "   FITNESS PLATFORM - LOAD TEST"
echo "========================================"
echo ""

# Check k6
echo "[1/3] Checking k6 installation..."
if command -v k6 &> /dev/null; then
    echo "  ✓ k6 is installed: $(k6 version)"
else
    echo "  ✗ k6 is not installed!"
    echo ""
    echo "Install k6:"
    echo "  Linux: curl https://dl.k6.io/k6-0.46.0-linux-amd64.tar.gz -L | tar xvz --strip-components 1"
    echo "  Then:  sudo mv k6 /usr/local/bin/"
    echo "  Or:    https://k6.io/docs/getting-started/installation/"
    exit 1
fi

# Check API
echo ""
echo "[2/3] Checking API availability..."
if curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/health" | grep -q "200"; then
    echo "  ✓ API is running at $BASE_URL"
else
    echo "  ✗ Cannot connect to $BASE_URL"
    echo "  Make sure services are running"
    exit 1
fi

# Run load test
echo ""
echo "[3/3] Starting load test..."
echo "  Base URL: $BASE_URL"
echo "  Duration: $DURATION"
echo "  VUs:      $VUS"
echo ""

export BASE_URL

mkdir -p logs

k6 run \
    --duration "$DURATION" \
    --vus "$VUS" \
    --out json=logs/load-test-results.json \
    scripts/load-test/load-test.js

echo ""
echo "========================================"
echo "   LOAD TEST COMPLETED!"
echo "========================================"
echo ""
echo "Results saved to: logs/load-test-results.json"
echo ""