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

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# API Base URL
API_URL="http://localhost:3000"

echo " Testing Payment API"
echo "======================"

# Test 1: Health Check
echo -e "\n${YELLOW}Test 1: Health Check${NC}"
curl -s "${API_URL}/health" | jq '.'

# Test 2: Create Payment
echo -e "\n${YELLOW}Test 2: Create Payment${NC}"
PAYMENT_ID=$(curl -s -X POST "${API_URL}/api/v1/payments" \
  -H "Content-Type: application/json" \
  -H "X-User-ID: user123" \
  -d '{
    "user_id": "user123",
    "amount": 100.50,
    "currency": "THB",
    "description": "Test payment"
  }' | jq -r '.payment_id')

echo "Payment ID: ${PAYMENT_ID}"

# Test 3: Get Payment Status
echo -e "\n${YELLOW}Test 3: Get Payment Status (immediately)${NC}"
curl -s "${API_URL}/api/v1/payments/${PAYMENT_ID}" | jq '.'

# Wait for processing
echo -e "\n${YELLOW}Waiting 3 seconds for processing...${NC}"
sleep 3

# Test 4: Get Payment Status (after processing)
echo -e "\n${YELLOW}Test 4: Get Payment Status (after processing)${NC}"
curl -s "${API_URL}/api/v1/payments/${PAYMENT_ID}" | jq '.'

# Test 5: Get Stats
echo -e "\n${YELLOW}Test 5: Get Payment Stats${NC}"
curl -s "${API_URL}/api/v1/payments/metrics/stats" | jq '.'

# Test 6: Rate Limiting
echo -e "\n${YELLOW}Test 6: Rate Limiting (send 15 requests rapidly)${NC}"
for i in {1..15}; do
  STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST "${API_URL}/api/v1/payments" \
    -H "Content-Type: application/json" \
    -H "X-User-ID: user123" \
    -d '{
      "user_id": "user123",
      "amount": 10,
      "currency": "THB",
      "description": "Rate limit test"
    }')

  if [ "$STATUS" == "429" ]; then
    echo -e "${RED}Request $i: Rate Limited (429)${NC}"
  else
    echo -e "${GREEN}Request $i: Success ($STATUS)${NC}"
  fi

  sleep 0.05
done

echo -e "\n${GREEN} All tests completed!${NC}"
