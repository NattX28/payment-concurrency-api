#!/bin/bash

set -e

command -v jq >/dev/null 2>&1 || {
  echo "❌ jq is required but not installed"
  echo ""
  echo "Install jq:"
  echo "  macOS   : brew install jq"
  echo "  Ubuntu  : sudo apt install jq"
  echo "  Windows : Download jq-win64.exe"
  echo "            https://jqlang.github.io/jq/download/"
  echo "            Rename to jq.exe and put it in:"
  echo "            C:\\Program Files\\Git\\usr\\bin\\"
  exit 1
}

command -v parallel >/dev/null 2>&1 || {
  echo "❌ GNU Parallel is required but not installed"
  echo ""
  echo "Install GNU Parallel:"
  echo "  macOS   : brew install parallel"
  echo "  Ubuntu  : sudo apt install parallel"
  echo "  Windows : Use WSL (Windows Subsystem for Linux)"
  exit 1
}

echo "Load Testing Payment API"
echo "========================"

API_URL="http://localhost:3000"
NUM_REQUESTS=1000

echo "Sending ${NUM_REQUESTS} concurrent requests..."

seq 1 ${NUM_REQUESTS} | parallel --will-cite -j 100 "curl -s --max-time 10 -X POST '${API_URL}/api/v1/payments' \
    -H 'Content-Type: application/json' \
    -H 'X-User-ID: user{}' \
    -d '{\"user_id\": \"user{}\", \"amount\":100, \"currency\": \"THB\", \"description\": \"Load test\"}' > /dev/null 2>&1"

echo ""
echo "✅ Load test complete!"
echo ""
echo "📊 Stats:"
curl -s "${API_URL}/api/v1/payments/metrics/stats" | jq '.'
