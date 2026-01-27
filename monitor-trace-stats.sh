#!/bin/bash

# Trace Generator 與 Tempo Anomaly Service 整合監控腳本
# 用途：監控 Trace Generator 產生的 traces 並查詢統計資料

set -e

# 顏色定義
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 配置參數
ANOMALY_SERVICE_URL="http://localhost:8081"
CHECK_INTERVAL=${CHECK_INTERVAL:-30}  # 每 30 秒檢查一次
MAX_CHECKS=${MAX_CHECKS:-5}  # 最多檢查 5 次

echo ""
echo "=========================================="
echo "  Trace Generator 統計監控"
echo "=========================================="
echo ""

# 步驟 1: 檢查服務狀態
echo -e "${BLUE}[步驟 1/4]${NC} 檢查服務狀態..."
echo ""

# 檢查 trace-demo-app
if docker ps | grep -q "trace-demo-app"; then
    echo -e "${GREEN}✓${NC} trace-demo-app 正在運行"
else
    echo -e "${RED}✗${NC} trace-demo-app 未運行"
    echo "請先啟動: docker-compose -f docker-compose-deploy.yml up -d"
    exit 1
fi

# 檢查 trace-generator
if docker ps | grep -q "trace-generator"; then
    echo -e "${GREEN}✓${NC} trace-generator 正在運行"
else
    echo -e "${YELLOW}!${NC} trace-generator 未運行，正在啟動..."
    docker-compose -f docker-compose-deploy.yml up -d trace-generator
    sleep 3
    echo -e "${GREEN}✓${NC} trace-generator 已啟動"
fi

# 檢查 tempo-anomaly-service
if docker ps | grep -q "tempo-anomaly-service"; then
    echo -e "${GREEN}✓${NC} tempo-anomaly-service 正在運行"
else
    echo -e "${RED}✗${NC} tempo-anomaly-service 未運行"
    echo "請參考 tempo-latency-anomaly-service 文檔啟動服務"
    exit 1
fi

# 檢查 tempo-anomaly-service 健康狀態
if curl -s -f "${ANOMALY_SERVICE_URL}/healthz" > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} tempo-anomaly-service 健康檢查通過 (${ANOMALY_SERVICE_URL})"
else
    echo -e "${RED}✗${NC} tempo-anomaly-service 無法訪問"
    exit 1
fi

echo ""

# 步驟 2: 顯示 Trace Generator 狀態
echo -e "${BLUE}[步驟 2/4]${NC} Trace Generator 運行狀態"
echo ""

# 從配置讀取間隔時間
interval=$(docker inspect trace-generator | jq -r '.[0].Config.Env[] | select(startswith("INTERVAL_SECONDS="))' | cut -d= -f2)
enabled_apis=$(docker inspect trace-generator | jq -r '.[0].Config.Env[] | select(startswith("ENABLED_APIS="))' | cut -d= -f2)

echo "配置資訊："
echo "  - 呼叫間隔: ${interval} 秒"
echo "  - 啟用的 API: ${enabled_apis}"
echo ""

echo "最近日誌（最後 10 行）："
echo "----------------------------------------"
docker logs --tail=10 trace-generator 2>&1
echo "----------------------------------------"
echo ""

# 步驟 3: 持續監控統計資料
echo -e "${BLUE}[步驟 3/4]${NC} 監控 Tempo Anomaly Service 統計資料"
echo ""
echo "將每 ${CHECK_INTERVAL} 秒檢查一次，共檢查 ${MAX_CHECKS} 次"
echo ""

for i in $(seq 1 $MAX_CHECKS); do
    echo -e "${CYAN}━━━ 檢查 #${i}/${MAX_CHECKS} ━━━${NC}"
    echo ""
    
    # 查詢可用服務
    response=$(curl -s "${ANOMALY_SERVICE_URL}/v1/available")
    total_services=$(echo "$response" | jq -r '.totalServices // 0')
    total_endpoints=$(echo "$response" | jq -r '.totalEndpoints // 0')
    
    echo "統計摘要："
    echo "  • 服務數: ${total_services}"
    echo "  • 端點數: ${total_endpoints}"
    echo ""
    
    if [ "$total_services" -gt 0 ]; then
        echo "服務詳情："
        echo "$response" | jq -r '.services[] | "  [\(.service)]\n    端點: \(.endpoint)\n    時間桶: \(.buckets | join(", "))\n"'
        
        # 顯示第一個服務的 baseline 資訊（簡化版本）
        echo "Baseline 範例（第一個端點）："
        first_service=$(echo "$response" | jq -r '.services[0].service')
        first_endpoint=$(echo "$response" | jq -r '.services[0].endpoint')
        first_bucket=$(echo "$response" | jq -r '.services[0].buckets[0]')
        
        if [ -n "$first_bucket" ] && [ "$first_bucket" != "null" ]; then
            hour=$(echo "$first_bucket" | cut -d'|' -f1)
            dayType=$(echo "$first_bucket" | cut -d'|' -f2)
            
            # URL encode endpoint
            endpoint_encoded=$(echo "$first_endpoint" | sed 's/ /%20/g' | sed 's/\//%2F/g')
            baseline=$(curl -s "${ANOMALY_SERVICE_URL}/v1/baseline?service=${first_service}&endpoint=${endpoint_encoded}&hour=${hour}&dayType=${dayType}")
            
            if echo "$baseline" | jq -e '.baseline' > /dev/null 2>&1; then
                p95=$(echo "$baseline" | jq -r '.baseline.p95')
                median=$(echo "$baseline" | jq -r '.baseline.median')
                sampleCount=$(echo "$baseline" | jq -r '.baseline.sampleCount')
                
                echo "  服務: $first_service"
                echo "  端點: $first_endpoint"
                echo "  時間: ${hour}:00 (${dayType})"
                echo "  P95: ${p95}ms | Median: ${median}ms | 樣本數: ${sampleCount}"
            else
                echo "  (無法取得 baseline 資料)"
            fi
        fi
        echo ""
    else
        echo -e "${YELLOW}尚無統計資料，繼續等待...${NC}"
        echo ""
    fi
    
    # 如果不是最後一次檢查，等待一段時間
    if [ $i -lt $MAX_CHECKS ]; then
        echo "等待 ${CHECK_INTERVAL} 秒後進行下一次檢查..."
        echo ""
        sleep $CHECK_INTERVAL
    fi
done

# 步驟 4: 最終報告
echo ""
echo -e "${BLUE}[步驟 4/4]${NC} 最終報告"
echo ""
echo "=========================================="

# 最後一次查詢
final_response=$(curl -s "${ANOMALY_SERVICE_URL}/v1/available")
final_total=$(echo "$final_response" | jq -r '.totalServices // 0')

if [ "$final_total" -gt 0 ]; then
    echo -e "${GREEN}✓ 成功收集到統計資料！${NC}"
    echo ""
    echo "完整回應 (JSON):"
    echo "$final_response" | jq '.'
else
    echo -e "${YELLOW}⚠ 仍未收集到足夠的統計資料${NC}"
    echo ""
    echo "可能原因："
    echo "  1. 需要更多時間讓 traces 被處理"
    echo "  2. 最小樣本數尚未達到"
    echo ""
    echo "建議："
    echo "  - 檢查 tempo-anomaly-service 日誌"
    echo "  - 確認 Tempo 正在收集 traces"
    echo "  - 增加等待時間重新執行"
fi

echo ""
echo "=========================================="
echo ""

# 提供有用的命令
echo "有用的命令："
echo ""
echo "1. 即時監控 Trace Generator:"
echo "   docker logs -f trace-generator"
echo ""
echo "2. 查看統計資料:"
echo "   curl ${ANOMALY_SERVICE_URL}/v1/available | jq ."
echo ""
echo "3. 查詢特定 baseline (範例):"
echo "   curl '${ANOMALY_SERVICE_URL}/v1/baseline?service=trace-demo-service&endpoint=POST%20%2Fapi%2Forder%2Fcreate&hour=17&dayType=weekday' | jq ."
echo ""
echo "4. 檢測異常 (範例):"
echo "   curl -X POST ${ANOMALY_SERVICE_URL}/v1/anomaly/check \\"
echo "     -H 'Content-Type: application/json' \\"
echo "     -d '{\"service\":\"trace-demo-service\",\"endpoint\":\"POST /api/order/create\",\"timestampNano\":'\$(date +%s)000000000',\"durationMs\":100}' | jq ."
echo ""
echo "5. 停止 Trace Generator:"
echo "   docker-compose -f docker-compose-deploy.yml stop trace-generator"
echo ""
echo "6. 重新執行監控（延長檢查時間）:"
echo "   CHECK_INTERVAL=60 MAX_CHECKS=10 ./monitor-trace-stats.sh"
echo ""

echo "監控完成！"
echo ""
