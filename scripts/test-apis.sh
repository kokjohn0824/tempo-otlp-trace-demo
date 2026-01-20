#!/bin/bash

# Tempo OTLP Trace Demo - API Test Script
# This script tests all API endpoints and generates various trace patterns

set -e

BASE_URL="${BASE_URL:-http://localhost:8080}"
SLEEP_BETWEEN_CALLS="${SLEEP_BETWEEN_CALLS:-2}"

echo "=========================================="
echo "Tempo OTLP Trace Demo - API Test Script"
echo "=========================================="
echo "Base URL: $BASE_URL"
echo ""

# Color codes for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print section header
print_section() {
    echo ""
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}"
}

# Function to make API call
call_api() {
    local method=$1
    local endpoint=$2
    local data=$3
    
    echo -e "${GREEN}Calling: $method $endpoint${NC}"
    
    if [ "$method" = "GET" ]; then
        curl -s -X GET "$BASE_URL$endpoint" | jq '.' || echo "Response received"
    else
        curl -s -X POST "$BASE_URL$endpoint" \
            -H "Content-Type: application/json" \
            -d "$data" | jq '.' || echo "Response received"
    fi
    
    echo ""
    sleep $SLEEP_BETWEEN_CALLS
}

# Health check
print_section "Health Check"
call_api "GET" "/health"

# Test 1: User Profile (Simple, Fast)
print_section "Test 1: User Profile (110-310ms, 4-5 spans)"
call_api "GET" "/api/user/profile?user_id=user_12345"

# Test 2: Search (Medium complexity)
print_section "Test 2: Search (210-530ms, 6-7 spans)"
call_api "GET" "/api/search?q=laptop&page=1&limit=10"

# Test 3: Order Creation (Complex)
print_section "Test 3: Order Creation (600-1500ms, 10-12 spans)"
call_api "POST" "/api/order/create" '{
  "user_id": "user_12345",
  "product_id": "prod_98765",
  "quantity": 2,
  "price": 299.99
}'

# Test 4: Batch Processing (Variable)
print_section "Test 4: Batch Processing (300-1500ms, 6-15 spans)"
call_api "POST" "/api/batch/process" '{
  "items": ["item1", "item2", "item3", "item4", "item5"]
}'

# Test 5: Report Generation (LONG TRACE - This is the key one!)
print_section "Test 5: Report Generation - LONG TRACE (1500-3500ms, 10-12 spans)"
echo -e "${YELLOW}⚠️  This will take 1.5-3.5 seconds - this is the 'abnormally long' trace!${NC}"
call_api "POST" "/api/report/generate" '{
  "report_type": "sales",
  "start_date": "2024-01-01",
  "end_date": "2024-01-31",
  "filters": ["region:US", "category:electronics"]
}'

# Test 6: Custom Simulation (Configurable)
print_section "Test 6: Custom Simulation - Shallow & Wide"
call_api "GET" "/api/simulate?depth=2&breadth=4&duration=50&variance=0.3"

# Test 7: Custom Simulation - Deep & Narrow
print_section "Test 7: Custom Simulation - Deep & Narrow"
call_api "GET" "/api/simulate?depth=5&breadth=2&duration=100&variance=0.5"

# Test 8: Mixed load simulation
print_section "Test 8: Mixed Load Simulation"
echo "Generating mixed traffic pattern..."

for i in {1..3}; do
    echo -e "${GREEN}Round $i of mixed calls${NC}"
    call_api "GET" "/api/user/profile?user_id=user_$i" &
    call_api "GET" "/api/search?q=test$i" &
    wait
    sleep 1
done

# Test 9: Generate multiple long traces for analysis
print_section "Test 9: Generate Multiple Long Traces"
echo "Generating 3 long traces for 'longest span' analysis..."

for i in {1..3}; do
    echo -e "${YELLOW}Long trace $i/3${NC}"
    call_api "POST" "/api/report/generate" '{
      "report_type": "analytics",
      "start_date": "2024-01-01",
      "end_date": "2024-12-31",
      "filters": ["all"]
    }'
done

# Summary
print_section "Test Complete!"
echo "All API endpoints have been tested."
echo ""
echo "Next steps:"
echo "1. Open Grafana at http://localhost:3000"
echo "2. Go to Explore → Select Tempo datasource"
echo "3. Search for traces by service name: trace-demo-service"
echo "4. Look for the longest traces (should be from /api/report/generate)"
echo ""
echo "Expected trace durations:"
echo "  - /api/user/profile:     110-310ms   (4-5 spans)"
echo "  - /api/search:           210-530ms   (6-7 spans)"
echo "  - /api/order/create:     600-1500ms  (10-12 spans)"
echo "  - /api/batch/process:    300-1500ms  (6-15 spans)"
echo "  - /api/report/generate:  1500-3500ms (10-12 spans) ⭐ LONGEST"
echo "  - /api/simulate:         variable     (configurable)"
echo ""
