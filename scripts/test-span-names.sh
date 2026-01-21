#!/bin/bash

# 測試 span-names API 的腳本

set -e

# 顏色定義
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

BASE_URL="${BASE_URL:-http://localhost:8080}"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  測試 Span Names API${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# 測試 1: 獲取所有 span names
echo -e "${YELLOW}測試 1: 獲取所有可用的 span names${NC}"
echo "GET ${BASE_URL}/api/span-names"
echo ""

RESPONSE=$(curl -s "${BASE_URL}/api/span-names")
echo "$RESPONSE" | jq '.'

# 檢查是否有資料
COUNT=$(echo "$RESPONSE" | jq -r '.count')
if [ "$COUNT" -gt 0 ]; then
    echo -e "${GREEN}✓ 成功獲取 ${COUNT} 個 span names${NC}"
else
    echo -e "${RED}✗ 未找到任何 span names${NC}"
    exit 1
fi

echo ""
echo -e "${BLUE}----------------------------------------${NC}"
echo ""

# 測試 2: 從結果中選擇第一個 span name 並查詢其原始碼
FIRST_SPAN_NAME=$(echo "$RESPONSE" | jq -r '.span_names[0].span_name')
echo -e "${YELLOW}測試 2: 使用第一個 span name 查詢原始碼${NC}"
echo "Span Name: ${FIRST_SPAN_NAME}"
echo "POST ${BASE_URL}/api/source-code"
echo ""

SOURCE_CODE_RESPONSE=$(curl -s -X POST "${BASE_URL}/api/source-code" \
    -H "Content-Type: application/json" \
    -d "{\"spanName\": \"${FIRST_SPAN_NAME}\"}")

echo "$SOURCE_CODE_RESPONSE" | jq '.'

# 檢查是否成功獲取原始碼
if echo "$SOURCE_CODE_RESPONSE" | jq -e '.source_code' > /dev/null 2>&1; then
    SOURCE_CODE_LENGTH=$(echo "$SOURCE_CODE_RESPONSE" | jq -r '.source_code' | wc -l)
    FILE_PATH=$(echo "$SOURCE_CODE_RESPONSE" | jq -r '.file_path')
    FUNCTION_NAME=$(echo "$SOURCE_CODE_RESPONSE" | jq -r '.function_name')
    
    echo -e "${GREEN}✓ 成功獲取原始碼${NC}"
    echo -e "  檔案路徑: ${FILE_PATH}"
    echo -e "  函數名稱: ${FUNCTION_NAME}"
    echo -e "  程式碼行數: ${SOURCE_CODE_LENGTH}"
else
    echo -e "${RED}✗ 無法獲取原始碼${NC}"
    exit 1
fi

echo ""
echo -e "${BLUE}----------------------------------------${NC}"
echo ""

# 測試 3: 顯示所有 span names 的摘要
echo -e "${YELLOW}測試 3: 所有可用的 span names 摘要${NC}"
echo ""
echo "$RESPONSE" | jq -r '.span_names[] | "- \(.span_name) (\(.function_name) in \(.file_path))"'

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  所有測試通過！${NC}"
echo -e "${GREEN}========================================${NC}"
