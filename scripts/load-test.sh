#!/bin/bash
# scripts/load-test.sh
# Fitness Platform - Load Testing Script for Linux (k6 wrapper)

BASE_URL=${1:-https://localhost:8443}
DURATION=${2:-2m}
VUS=${3:-50}

echo "========================================"
echo "   LOAD TESTING (k6)"
echo "========================================"
echo "Base URL: $BASE_URL"
echo "Duration: $DURATION"
echo "VUs:      $VUS"
echo ""

# Check if k6 is installed
echo "[1/3] Checking k6 installation..."
if command -v k6 &> /dev/null; then
    echo "  ✓ k6 is installed: $(k6 version)"
else
    echo "  ✗ k6 is not installed!"
    echo "     Install from: https://k6.io/docs/getting-started/installation/"
    echo "     Linux: curl https://dl.k6.io/k6-0.46.0-linux-amd64.tar.gz -L | tar xvz --strip-components 1"
    echo "     Then: sudo mv k6 /usr/local/bin/"
    exit 1
fi

# Run load test
echo ""
echo "[2/3] Running load test..."
echo "  Script: scripts/load-test/load-test.js"
echo "  Command: k6 run --duration $DURATION --vus $VUS scripts/load-test/load-test.js"
echo ""

export BASE_URL
k6 run --duration $DURATION --vus $VUS scripts/load-test/load-test.js

if [ $? -eq 0 ]; then
    echo ""
    echo "[3/3] Load test completed successfully!"
else
    echo ""
    echo "[3/3] Load test completed with errors!"
fi

echo ""
echo "========================================"
echo "   LOAD TEST SUMMARY"
echo "========================================"
echo "Check logs above for detailed metrics."
echo "========================================"

exit $?