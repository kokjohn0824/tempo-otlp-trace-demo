#!/bin/bash

# Example workflow for using Source Code Analysis API
# This script demonstrates a complete workflow from generating traces to analyzing them

set -e

BASE_URL="http://localhost:8080"
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "=========================================="
echo "Source Code Analysis - Example Workflow"
echo "=========================================="
echo ""

# Step 1: Generate a trace
echo -e "${BLUE}Step 1: Generating a test trace...${NC}"
ORDER_RESPONSE=$(curl -s -X POST "${BASE_URL}/api/order/create" \
    -H "Content-Type: application/json" \
    -d '{
        "user_id": "demo_user_123",
        "product_id": "demo_product_456",
        "quantity": 3,
        "price": 149.99
    }')

ORDER_ID=$(echo "$ORDER_RESPONSE" | jq -r '.order_id')
echo -e "${GREEN}✓ Order created: $ORDER_ID${NC}"
echo ""

# Step 2: Wait for trace to be available
echo -e "${BLUE}Step 2: Waiting for trace to be available in Tempo...${NC}"
echo -e "${YELLOW}(In a real scenario, you would get trace_id and span_id from Grafana)${NC}"
sleep 3
echo ""

# Step 3: Show how to query mappings
echo -e "${BLUE}Step 3: Checking available source code mappings...${NC}"
MAPPINGS=$(curl -s "${BASE_URL}/api/mappings")
MAPPING_COUNT=$(echo "$MAPPINGS" | jq '.mappings | length')
echo -e "${GREEN}✓ Found $MAPPING_COUNT source code mappings${NC}"
echo ""
echo "Sample mappings:"
echo "$MAPPINGS" | jq '.mappings[0:3] | .[] | {span_name, function_name, file_path}'
echo ""

# Step 4: Explain the manual steps
echo -e "${BLUE}Step 4: Manual steps to analyze the trace:${NC}"
echo ""
echo "1. Open Grafana: http://localhost:3000"
echo "2. Go to Explore → Tempo"
echo "3. Search for traces with service.name=\"trace-demo-service\""
echo "4. Find the trace for your order (look for operation: POST /api/order/create)"
echo "5. Click on the trace to view details"
echo "6. Copy the Trace ID and Span ID from the UI"
echo ""

# Step 5: Show example API call
echo -e "${BLUE}Step 5: Example API call (with placeholder IDs):${NC}"
echo ""
echo "Once you have the trace_id and span_id, run:"
echo ""
echo -e "${YELLOW}curl \"${BASE_URL}/api/source-code?span_id=YOUR_SPAN_ID&trace_id=YOUR_TRACE_ID\" | jq .${NC}"
echo ""

# Step 6: Show example response structure
echo -e "${BLUE}Step 6: Expected response structure:${NC}"
echo ""
cat << 'EOF'
{
  "span_id": "abc123...",
  "span_name": "POST /api/order/create",
  "trace_id": "xyz789...",
  "duration": "1.23s",
  "file_path": "handlers/order.go",
  "function_name": "CreateOrder",
  "start_line": 21,
  "end_line": 85,
  "source_code": "func CreateOrder(...) { ... }",
  "attributes": {
    "http.method": "POST",
    "user.id": "demo_user_123"
  },
  "child_spans": [
    {
      "span_id": "def456",
      "span_name": "validateOrder",
      "duration": "52.3ms",
      "function_name": "validateOrder"
    },
    {
      "span_id": "ghi789",
      "span_name": "processPayment",
      "duration": "850.2ms",
      "function_name": "processPayment"
    }
  ]
}
EOF
echo ""

# Step 7: LLM Analysis prompt
echo -e "${BLUE}Step 7: Using the data with LLM:${NC}"
echo ""
echo "Copy the JSON response and use this prompt with your LLM:"
echo ""
cat << 'EOF'
---
我有一個 API 的效能問題。以下是從 OpenTelemetry trace 獲取的資訊：

[貼上 JSON 回應]

請分析：
1. 主要的效能瓶頸在哪裡？
2. 哪些子操作佔用最多時間？
3. 根據原始碼，有哪些可能的優化方案？
4. 建議的優先順序是什麼？
---
EOF
echo ""

# Step 8: Additional tips
echo -e "${BLUE}Step 8: Additional tips:${NC}"
echo ""
echo "• To analyze child spans, use their span_id with the same trace_id"
echo "• Compare multiple traces to identify patterns"
echo "• Use the duration field to prioritize optimization efforts"
echo "• Check the attributes for additional context (database queries, external APIs, etc.)"
echo ""

echo "=========================================="
echo "Workflow Complete!"
echo "=========================================="
echo ""
echo "For more examples, see:"
echo "• USAGE_EXAMPLE.md - Detailed usage examples"
echo "• SOURCE_CODE_API.md - Complete API documentation"
echo ""
