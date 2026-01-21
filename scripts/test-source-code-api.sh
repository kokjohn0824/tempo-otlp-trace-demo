#!/bin/bash

# Test script for Source Code Analysis API
# This script tests all the new API endpoints

set -e

BASE_URL="http://localhost:8080"
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "=========================================="
echo "Source Code Analysis API Test Script"
echo "=========================================="
echo ""

# Function to print colored output
print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}ℹ $1${NC}"
}

# Test 1: Check if server is running
echo "Test 1: Checking if server is running..."
if curl -s "${BASE_URL}/health" > /dev/null; then
    print_success "Server is running"
else
    print_error "Server is not running. Please start the server first."
    exit 1
fi
echo ""

# Test 2: Get all mappings
echo "Test 2: Getting all source code mappings..."
MAPPINGS_RESPONSE=$(curl -s "${BASE_URL}/api/mappings")
MAPPINGS_COUNT=$(echo "$MAPPINGS_RESPONSE" | jq '.mappings | length')
if [ "$MAPPINGS_COUNT" -gt 0 ]; then
    print_success "Found $MAPPINGS_COUNT mappings"
    echo "Sample mapping:"
    echo "$MAPPINGS_RESPONSE" | jq '.mappings[0]'
else
    print_error "No mappings found"
fi
echo ""

# Test 3: Generate a trace by calling an API
echo "Test 3: Generating a trace by calling /api/order/create..."
ORDER_RESPONSE=$(curl -s -X POST "${BASE_URL}/api/order/create" \
    -H "Content-Type: application/json" \
    -d '{
        "user_id": "test_user_123",
        "product_id": "product_456",
        "quantity": 2,
        "price": 99.99
    }')

if echo "$ORDER_RESPONSE" | jq -e '.order_id' > /dev/null; then
    ORDER_ID=$(echo "$ORDER_RESPONSE" | jq -r '.order_id')
    print_success "Order created: $ORDER_ID"
else
    print_error "Failed to create order"
    echo "$ORDER_RESPONSE"
fi
echo ""

# Wait for trace to be available in Tempo
print_info "Waiting 5 seconds for trace to be available in Tempo..."
sleep 5
echo ""

# Test 4: Query Tempo for recent traces (requires Tempo API)
echo "Test 4: Querying Tempo for recent traces..."
TEMPO_URL="${TEMPO_URL:-http://localhost:3200}"
print_info "Using Tempo URL: $TEMPO_URL"

# Try to search for recent traces
TEMPO_SEARCH=$(curl -s "${TEMPO_URL}/api/search?tags=service.name=trace-demo-service&limit=5" || echo "{}")

if echo "$TEMPO_SEARCH" | jq -e '.traces' > /dev/null 2>&1; then
    TRACE_COUNT=$(echo "$TEMPO_SEARCH" | jq '.traces | length')
    print_success "Found $TRACE_COUNT recent traces in Tempo"
    
    if [ "$TRACE_COUNT" -gt 0 ]; then
        # Get the first trace ID
        TRACE_ID=$(echo "$TEMPO_SEARCH" | jq -r '.traces[0].traceID')
        print_info "Using trace ID: $TRACE_ID"
        
        # Get the full trace details
        TRACE_DETAILS=$(curl -s "${TEMPO_URL}/api/traces/${TRACE_ID}")
        
        if echo "$TRACE_DETAILS" | jq -e '.batches' > /dev/null 2>&1; then
            # Extract span IDs
            SPAN_ID=$(echo "$TRACE_DETAILS" | jq -r '.batches[0].scopeSpans[0].spans[0].spanId' | base64 -d | xxd -p -c 16)
            print_info "Found span ID: $SPAN_ID"
            
            # Test 5: Get source code for the span
            echo ""
            echo "Test 5: Getting source code for span..."
            SOURCE_CODE_RESPONSE=$(curl -s "${BASE_URL}/api/source-code?span_id=${SPAN_ID}&trace_id=${TRACE_ID}")
            
            if echo "$SOURCE_CODE_RESPONSE" | jq -e '.source_code' > /dev/null 2>&1; then
                print_success "Successfully retrieved source code"
                echo "Response summary:"
                echo "$SOURCE_CODE_RESPONSE" | jq '{
                    span_name: .span_name,
                    duration: .duration,
                    file_path: .file_path,
                    function_name: .function_name,
                    child_spans_count: (.child_spans | length),
                    source_code_lines: (.source_code | split("\n") | length)
                }'
            else
                print_error "Failed to retrieve source code"
                echo "$SOURCE_CODE_RESPONSE" | jq '.'
            fi
        else
            print_info "Trace format not as expected (might be using different Tempo version)"
        fi
    fi
else
    print_info "Could not query Tempo search API (this is optional)"
    print_info "You can manually test with: curl '${BASE_URL}/api/source-code?span_id=XXX&trace_id=YYY'"
fi
echo ""

# Test 6: Add a new mapping
echo "Test 6: Adding a new test mapping..."
NEW_MAPPING_RESPONSE=$(curl -s -X POST "${BASE_URL}/api/mappings" \
    -H "Content-Type: application/json" \
    -d '{
        "mappings": [
            {
                "span_name": "test_operation",
                "file_path": "handlers/order.go",
                "function_name": "TestFunction",
                "start_line": 1,
                "end_line": 10,
                "description": "Test mapping for demonstration"
            }
        ]
    }')

if echo "$NEW_MAPPING_RESPONSE" | jq -e '.status' > /dev/null; then
    STATUS=$(echo "$NEW_MAPPING_RESPONSE" | jq -r '.status')
    if [ "$STATUS" = "success" ]; then
        print_success "Successfully added new mapping"
    else
        print_error "Failed to add mapping"
    fi
else
    print_error "Invalid response when adding mapping"
fi
echo ""

# Test 7: Verify the new mapping was added
echo "Test 7: Verifying new mapping..."
UPDATED_MAPPINGS=$(curl -s "${BASE_URL}/api/mappings")
if echo "$UPDATED_MAPPINGS" | jq -e '.mappings[] | select(.span_name == "test_operation")' > /dev/null; then
    print_success "New mapping verified in the list"
else
    print_error "New mapping not found"
fi
echo ""

# Test 8: Delete the test mapping
echo "Test 8: Deleting test mapping..."
DELETE_RESPONSE=$(curl -s -X DELETE "${BASE_URL}/api/mappings?span_name=test_operation")
if echo "$DELETE_RESPONSE" | jq -e '.status' > /dev/null; then
    STATUS=$(echo "$DELETE_RESPONSE" | jq -r '.status')
    if [ "$STATUS" = "success" ]; then
        print_success "Successfully deleted test mapping"
    else
        print_error "Failed to delete mapping"
    fi
else
    print_error "Invalid response when deleting mapping"
fi
echo ""

# Test 9: Reload mappings from file
echo "Test 9: Reloading mappings from file..."
RELOAD_RESPONSE=$(curl -s -X POST "${BASE_URL}/api/mappings/reload")
if echo "$RELOAD_RESPONSE" | jq -e '.status' > /dev/null; then
    STATUS=$(echo "$RELOAD_RESPONSE" | jq -r '.status')
    COUNT=$(echo "$RELOAD_RESPONSE" | jq -r '.count')
    if [ "$STATUS" = "success" ]; then
        print_success "Successfully reloaded $COUNT mappings"
    else
        print_error "Failed to reload mappings"
    fi
else
    print_error "Invalid response when reloading mappings"
fi
echo ""

# Test 10: Test error handling - missing parameters
echo "Test 10: Testing error handling (missing parameters)..."
ERROR_RESPONSE=$(curl -s "${BASE_URL}/api/source-code")
if echo "$ERROR_RESPONSE" | grep -q "Missing required parameters"; then
    print_success "Error handling works correctly"
else
    print_error "Error handling not working as expected"
fi
echo ""

echo "=========================================="
echo "Test Summary"
echo "=========================================="
print_info "All basic tests completed!"
print_info ""
print_info "Manual testing steps:"
print_info "1. Call any API endpoint to generate traces"
print_info "2. View traces in Grafana (http://localhost:3000)"
print_info "3. Copy trace ID and span ID from Grafana"
print_info "4. Call: curl '${BASE_URL}/api/source-code?span_id=XXX&trace_id=YYY'"
print_info ""
print_info "For LLM analysis, you can pipe the output to jq:"
print_info "curl '${BASE_URL}/api/source-code?span_id=XXX&trace_id=YYY' | jq ."
echo ""
