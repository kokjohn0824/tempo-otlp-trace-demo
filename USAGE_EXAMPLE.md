# Source Code Analysis API - 使用範例

這份文件提供了詳細的使用範例，展示如何使用 Source Code Analysis API 來分析效能問題。

## 快速開始

### 1. 啟動服務

```bash
# 啟動所有服務（包括 Tempo, Grafana）
docker-compose up -d

# 啟動應用程式
go run main.go
```

### 2. 產生測試資料

```bash
# 呼叫 API 產生 traces
curl -X POST http://localhost:8080/api/order/create \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user_123",
    "product_id": "product_456",
    "quantity": 2,
    "price": 99.99
  }'
```

### 3. 在 Grafana 中查看 Trace

1. 開啟瀏覽器訪問 http://localhost:3000
2. 進入 Explore 頁面
3. 選擇 Tempo 資料源
4. 搜尋最近的 traces
5. 找到一個 duration 較長的 trace
6. 複製 **Trace ID** 和 **Span ID**

### 4. 獲取原始碼

```bash
# 使用從 Grafana 複製的 trace ID 和 span ID
curl "http://localhost:8080/api/source-code?span_id=YOUR_SPAN_ID&trace_id=YOUR_TRACE_ID" | jq .
```

## 完整範例：分析慢速 API

### 場景：訂單建立 API 太慢

#### 步驟 1: 產生測試訂單

```bash
curl -X POST http://localhost:8080/api/order/create \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user_123",
    "product_id": "product_456",
    "quantity": 5,
    "price": 199.99
  }'
```

**回應範例：**
```json
{
  "order_id": "orders_1234567890",
  "status": "success",
  "total_cost": 999.95,
  "message": "Order created successfully"
}
```

#### 步驟 2: 在 Grafana 找到 Trace

1. 訪問 http://localhost:3000
2. 進入 Explore → Tempo
3. 搜尋 `service.name="trace-demo-service"`
4. 找到 operation name 為 `POST /api/order/create` 的 trace
5. 點擊查看詳細資訊

**假設我們找到：**
- Trace ID: `a1b2c3d4e5f6g7h8`
- Root Span ID: `1234567890abcdef`
- Total Duration: `1.2s`

#### 步驟 3: 獲取原始碼和分析資料

```bash
curl "http://localhost:8080/api/source-code?span_id=1234567890abcdef&trace_id=a1b2c3d4e5f6g7h8" \
  | jq . > analysis_data.json
```

**回應範例：**
```json
{
  "span_id": "1234567890abcdef",
  "span_name": "POST /api/order/create",
  "trace_id": "a1b2c3d4e5f6g7h8",
  "duration": "1.20s",
  "file_path": "handlers/order.go",
  "function_name": "CreateOrder",
  "start_line": 21,
  "end_line": 85,
  "source_code": "func CreateOrder(w http.ResponseWriter, r *http.Request) {\n\tctx := r.Context()\n\t...\n}",
  "attributes": {
    "http.method": "POST",
    "http.route": "/api/order/create",
    "user.id": "user_123",
    "product.id": "product_456",
    "order.quantity": "5"
  },
  "child_spans": [
    {
      "span_id": "abc123",
      "span_name": "validateOrder",
      "duration": "52.3ms",
      "function_name": "validateOrder"
    },
    {
      "span_id": "def456",
      "span_name": "checkInventory",
      "duration": "105.7ms",
      "function_name": "checkInventory"
    },
    {
      "span_id": "ghi789",
      "span_name": "processPayment",
      "duration": "850.2ms",
      "function_name": "processPayment"
    },
    {
      "span_id": "jkl012",
      "span_name": "createShipment",
      "duration": "95.4ms",
      "function_name": "createShipment"
    }
  ]
}
```

#### 步驟 4: 分析結果

從上面的資料可以看出：

1. **總 Duration**: 1.20s
2. **最慢的子操作**: `processPayment` (850.2ms，佔 70.8%)
3. **其他子操作**: 相對較快

#### 步驟 5: 深入分析 processPayment

```bash
# 獲取 processPayment 的原始碼
curl "http://localhost:8080/api/source-code?span_id=ghi789&trace_id=a1b2c3d4e5f6g7h8" \
  | jq . > payment_analysis.json
```

**回應可能顯示：**
```json
{
  "span_id": "ghi789",
  "span_name": "processPayment",
  "duration": "850.2ms",
  "file_path": "handlers/order.go",
  "function_name": "processPayment",
  "source_code": "func processPayment(...) {\n\t// 呼叫外部 payment gateway\n\tcallPaymentGateway(...)\n\t...\n}",
  "child_spans": [
    {
      "span_id": "xyz789",
      "span_name": "callPaymentGateway",
      "duration": "800.5ms",
      "function_name": "callPaymentGateway"
    },
    {
      "span_id": "uvw456",
      "span_name": "recordTransaction",
      "duration": "25.3ms",
      "function_name": "recordTransaction"
    }
  ]
}
```

**分析結論：**
- `callPaymentGateway` 佔用了 800.5ms
- 這是一個外部 API 呼叫，可能的優化方案：
  1. 增加 timeout 設定
  2. 實作 retry 機制
  3. 考慮使用非同步處理
  4. 檢查網路連線品質

#### 步驟 6: 使用 LLM 分析

將 `analysis_data.json` 的內容提供給 LLM：

**Prompt 範例：**
```
我有一個訂單建立 API 的效能問題。以下是從 OpenTelemetry trace 和原始碼獲取的資訊：

[貼上 analysis_data.json 的內容]

請分析：
1. 主要的效能瓶頸在哪裡？
2. 為什麼 processPayment 這麼慢？
3. 有哪些可能的優化方案？
4. 優先順序建議？
```

**LLM 可能的回應：**
```
根據提供的資料分析：

1. **主要效能瓶頸**：
   - processPayment 佔用 850.2ms（70.8%）
   - 其中 callPaymentGateway 佔用 800.5ms（94%）

2. **慢速原因**：
   - 外部 payment gateway API 呼叫延遲高
   - 可能是網路延遲或第三方服務回應慢

3. **優化方案**：
   a. 短期優化：
      - 設定合理的 timeout（例如 5 秒）
      - 加入 circuit breaker 避免級聯失敗
      - 實作 retry 機制（最多 3 次）
   
   b. 中期優化：
      - 考慮非同步處理支付
      - 使用 webhook 接收支付結果
      - 先建立訂單，支付成功後更新狀態
   
   c. 長期優化：
      - 評估更換更快的 payment gateway
      - 實作本地快取減少重複查詢
      - 使用 CDN 或就近的 API endpoint

4. **優先順序**：
   1. 立即：加入 timeout 和 error handling
   2. 本週：實作 circuit breaker
   3. 本月：考慮非同步處理架構
```

## 其他使用場景

### 場景 1: 比較不同 API 的效能

```bash
# 測試多個 API
for api in "order/create" "user/profile" "report/generate"; do
  echo "Testing /api/$api"
  curl -X POST "http://localhost:8080/api/$api" \
    -H "Content-Type: application/json" \
    -d '{}' > /dev/null 2>&1
  sleep 1
done

# 在 Grafana 中比較它們的 duration
```

### 場景 2: 監控特定操作

```bash
# 建立監控腳本
#!/bin/bash
while true; do
  # 呼叫 API
  curl -X POST http://localhost:8080/api/order/create \
    -H "Content-Type: application/json" \
    -d '{"user_id":"user_123","product_id":"prod_456","quantity":1,"price":99.99}' \
    > /dev/null 2>&1
  
  sleep 10
done
```

### 場景 3: 自動化效能報告

```bash
#!/bin/bash
# 產生效能報告

# 1. 產生測試資料
echo "Generating test data..."
for i in {1..10}; do
  curl -X POST http://localhost:8080/api/order/create \
    -H "Content-Type: application/json" \
    -d "{\"user_id\":\"user_$i\",\"product_id\":\"prod_123\",\"quantity\":$i,\"price\":99.99}" \
    > /dev/null 2>&1
  sleep 2
done

# 2. 等待 traces 可用
sleep 10

# 3. 從 Grafana/Tempo 獲取 trace IDs（需要額外的腳本）
# 4. 對每個 trace 呼叫 source code API
# 5. 生成報告
```

## 管理映射表

### 查看所有映射

```bash
curl http://localhost:8080/api/mappings | jq '.mappings[] | {span_name, function_name, file_path}'
```

### 新增自訂映射

```bash
curl -X POST http://localhost:8080/api/mappings \
  -H "Content-Type: application/json" \
  -d '{
    "mappings": [
      {
        "span_name": "customOperation",
        "file_path": "handlers/custom.go",
        "function_name": "CustomHandler",
        "start_line": 15,
        "end_line": 80,
        "description": "Custom operation for special processing"
      }
    ]
  }'
```

### 批量更新映射

```bash
# 編輯 source_code_mappings.json
vim source_code_mappings.json

# 重新載入
curl -X POST http://localhost:8080/api/mappings/reload
```

### 刪除映射

```bash
curl -X DELETE "http://localhost:8080/api/mappings?span_name=customOperation"
```

## 整合到工作流程

### 1. 開發階段

每次新增或修改 handler 時：

```bash
# 1. 寫程式碼
vim handlers/new_feature.go

# 2. 更新映射
curl -X POST http://localhost:8080/api/mappings \
  -H "Content-Type: application/json" \
  -d '{
    "mappings": [
      {
        "span_name": "POST /api/new-feature",
        "file_path": "handlers/new_feature.go",
        "function_name": "NewFeatureHandler",
        "start_line": 10,
        "end_line": 50,
        "description": "New feature implementation"
      }
    ]
  }'

# 3. 測試
curl -X POST http://localhost:8080/api/new-feature
```

### 2. 測試階段

```bash
# 執行測試腳本
./scripts/test-source-code-api.sh
```

### 3. 生產環境監控

```bash
# 定期檢查慢速 traces
# 自動呼叫 source code API
# 發送警報或生成報告
```

## 故障排除

### 問題 1: 找不到 span

```bash
# 檢查 trace ID 和 span ID 是否正確
curl "http://localhost:3200/api/traces/YOUR_TRACE_ID" | jq .
```

### 問題 2: 沒有映射

```bash
# 檢查映射表
curl http://localhost:8080/api/mappings | jq '.mappings[] | select(.span_name == "YOUR_SPAN_NAME")'

# 如果沒有，新增映射
curl -X POST http://localhost:8080/api/mappings -H "Content-Type: application/json" -d '...'
```

### 問題 3: 無法連接 Tempo

```bash
# 檢查 Tempo 狀態
docker-compose ps tempo

# 檢查 Tempo API
curl http://localhost:3200/api/search

# 設定環境變數
export TEMPO_URL=http://localhost:3200
```

## 進階使用

### 使用 jq 過濾資料

```bash
# 只顯示 child spans 的 duration
curl "http://localhost:8080/api/source-code?span_id=XXX&trace_id=YYY" \
  | jq '.child_spans[] | {span_name, duration}'

# 找出最慢的 child span
curl "http://localhost:8080/api/source-code?span_id=XXX&trace_id=YYY" \
  | jq '.child_spans | sort_by(.duration) | reverse | .[0]'
```

### 整合到 CI/CD

```yaml
# .github/workflows/performance-test.yml
name: Performance Test

on: [push]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Start services
        run: docker-compose up -d
      - name: Run performance tests
        run: ./scripts/test-source-code-api.sh
      - name: Analyze results
        run: |
          # 呼叫 source code API
          # 分析結果
          # 如果 duration 超過閾值，失敗
```

## 總結

這個 Source Code Analysis API 提供了強大的工具來：

1. ✅ 快速定位效能瓶頸
2. ✅ 自動化效能分析流程
3. ✅ 整合 LLM 進行智能分析
4. ✅ 維護程式碼與 tracing 的對應關係
5. ✅ 生成詳細的效能報告

透過這些範例，您可以開始使用這個 API 來改善應用程式的效能！

## 📚 相關文件

- **[SOURCE_CODE_API.md](SOURCE_CODE_API.md)** - 完整 API 文件
- **[README.md](README.md)** - 專案說明
- **[CHANGELOG.md](CHANGELOG.md)** - 變更日誌
