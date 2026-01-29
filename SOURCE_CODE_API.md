# Source Code Analysis API

這個 API 允許您根據 Tempo 中的 span 資訊來獲取對應的原始碼，以供 LLM 分析 duration 較長的原因。

## 功能概述

1. **查詢原始碼**: 根據 span ID 和 trace ID 從 Tempo 查詢 span 資訊，並回傳對應的原始碼
2. **管理映射表**: 維護 span name 與原始碼位置的對應關係
3. **自動載入**: 系統啟動時自動載入映射表
4. **Swagger UI**: 提供互動式 API 文檔和測試介面

## Swagger UI

訪問 Swagger UI：http://localhost:8080/swagger/

```bash
# 生成 Swagger 文檔
make swagger

# 在瀏覽器開啟
make open-swagger
```

## API Endpoints

### 1. 獲取原始碼

根據 span ID 和 trace ID 獲取對應的原始碼及相關資訊。

**請求:**
```
GET /api/source-code?span_id={spanId}&trace_id={traceId}
```

**參數:**
- `span_id` (必填): Span ID
- `trace_id` (必填): Trace ID

**回應範例:**
```json
{
  "span_id": "abc123",
  "span_name": "POST /api/order/create",
  "trace_id": "xyz789",
  "duration": "1.23s",
  "file_path": "handlers/order.go",
  "function_name": "CreateOrder",
  "start_line": 21,
  "end_line": 85,
  "source_code": "func CreateOrder(w http.ResponseWriter, r *http.Request) {\n\t...\n}",
  "attributes": {
    "http.method": "POST",
    "http.route": "/api/order/create",
    "user.id": "user_123"
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
      "duration": "350.5ms",
      "function_name": "processPayment"
    }
  ]
}
```

**使用範例:**
```bash
curl "http://localhost:8080/api/source-code?span_id=abc123&trace_id=xyz789"
```

### 2. 查詢所有映射

獲取所有已設定的 span name 與原始碼位置的映射。

**請求:**
```
GET /api/mappings
```

**回應範例:**
```json
{
  "mappings": [
    {
      "span_name": "POST /api/order/create",
      "file_path": "handlers/order.go",
      "function_name": "CreateOrder",
      "start_line": 21,
      "end_line": 85,
      "description": "Handles order creation with comprehensive tracing"
    }
  ]
}
```

**使用範例:**
```bash
curl http://localhost:8080/api/mappings
```

### 3. 更新映射

新增或更新 span name 與原始碼位置的映射。

**請求:**
```
POST /api/mappings
Content-Type: application/json
```

**請求 Body:**
```json
{
  "mappings": [
    {
      "span_name": "customOperation",
      "file_path": "handlers/custom.go",
      "function_name": "CustomHandler",
      "start_line": 10,
      "end_line": 50,
      "description": "Custom operation handler"
    }
  ]
}
```

**回應範例:**
```json
{
  "status": "success",
  "message": "Mappings updated successfully",
  "count": 1
}
```

**使用範例:**
```bash
curl -X POST http://localhost:8080/api/mappings \
  -H "Content-Type: application/json" \
  -d '{
    "mappings": [
      {
        "span_name": "customOperation",
        "file_path": "handlers/custom.go",
        "function_name": "CustomHandler",
        "start_line": 10,
        "end_line": 50,
        "description": "Custom operation handler"
      }
    ]
  }'
```

### 4. 刪除映射

刪除指定的 span name 映射。

**請求:**
```
DELETE /api/mappings?span_name={spanName}
```

**參數:**
- `span_name` (必填): 要刪除的 span name

**回應範例:**
```json
{
  "status": "success",
  "message": "Mapping for 'customOperation' deleted successfully",
  "count": 1
}
```

**使用範例:**
```bash
curl -X DELETE "http://localhost:8080/api/mappings?span_name=customOperation"
```

### 5. 重新載入映射

從 `source_code_mappings.json` 檔案重新載入映射表。

**請求:**
```
POST /api/mappings/reload
```

**回應範例:**
```json
{
  "status": "success",
  "message": "Mappings reloaded successfully",
  "count": 42
}
```

**使用範例:**
```bash
curl -X POST http://localhost:8080/api/mappings/reload
```

## 映射表管理流程

### 方式 1: 透過 API 管理（推薦用於動態更新）

1. **新增映射**: 使用 `POST /api/mappings` 新增新的映射
2. **查看映射**: 使用 `GET /api/mappings` 查看當前所有映射
3. **刪除映射**: 使用 `DELETE /api/mappings` 刪除不需要的映射

### 方式 2: 直接編輯檔案（推薦用於批量更新）

1. 編輯 `source_code_mappings.json` 檔案
2. 呼叫 `POST /api/mappings/reload` 重新載入

## 映射表檔案格式

`source_code_mappings.json` 檔案格式：

```json
{
  "mappings": [
    {
      "span_name": "操作名稱（與 OpenTelemetry span 的 operation name 對應）",
      "file_path": "相對於專案根目錄的檔案路徑",
      "function_name": "函數名稱",
      "start_line": 起始行號,
      "end_line": 結束行號,
      "description": "可選的描述"
    }
  ]
}
```

## 工作流程

### 完整的使用流程

1. **產生 Trace**: 呼叫任何 API endpoint (例如 `/api/order/create`)
2. **查詢 Grafana**: 在 Grafana Tempo 中找到 trace，記下 trace ID 和 span ID
3. **獲取原始碼**: 呼叫 `/api/source-code?span_id=xxx&trace_id=yyy`
4. **分析結果**: 將回傳的原始碼、duration、child spans 等資訊提供給 LLM 分析

### LLM 分析範例

將 API 回應提供給 LLM，並詢問：

```
這個 API 的 duration 是 1.23s，比預期長。以下是原始碼和子 span 資訊：

[貼上 API 回應的 JSON]

請分析可能導致 duration 較長的原因，並提供優化建議。
```

LLM 可以根據：
- 原始碼邏輯
- Child spans 的 duration 分布
- Span attributes（例如資料庫查詢、外部 API 呼叫）

來分析效能瓶頸。

## 環境變數

### TEMPO_URL

設定 Tempo 查詢 API 的 URL。

**預設值**: `http://localhost:3200`

**範例**:
```bash
export TEMPO_URL=http://tempo:3200
```

## 錯誤處理

### 常見錯誤

1. **Missing required parameters**: 缺少 `span_id` 或 `trace_id` 參數
   - HTTP 400: "Missing required parameters: span_id and trace_id"

2. **Failed to query Tempo**: 無法連接到 Tempo 或查詢失敗
   - HTTP 500: "Failed to query Tempo: [error details]"
   - 檢查 `TEMPO_URL` 環境變數是否正確
   - 確認 Tempo 服務正在運行

3. **Span not found**: 在 trace 中找不到指定的 span
   - HTTP 404: "Span not found in trace"
   - 確認 span ID 和 trace ID 是否正確

4. **No source code mapping found**: 沒有找到對應的原始碼映射
   - HTTP 404: "No source code mapping found for span: [span name]"
   - 使用 `POST /api/mappings` 新增映射

5. **Failed to read source code**: 無法讀取原始碼檔案
   - HTTP 500: "Failed to read source code: [error details]"
   - 檢查檔案路徑是否正確
   - 確認檔案存在且有讀取權限

## 最佳實踐

### 1. 維護映射表

- 每次新增或修改 handler 時，同步更新映射表
- 使用有意義的 description 來說明每個函數的用途
- 定期檢查映射表的完整性

### 2. 行號管理

- 映射表中的 `start_line` 和 `end_line` 應該包含完整的函數定義
- 當修改程式碼時，記得更新對應的行號
- 可以考慮建立自動化腳本來掃描程式碼並更新行號

### 3. 效能考量

- 映射表在記憶體中維護，查詢速度快
- 原始碼檔案按需讀取，不會佔用大量記憶體
- Tempo 查詢可能需要幾秒鐘，請設定適當的 timeout

### 4. 安全性

- 此 API 會暴露原始碼，請確保適當的存取控制
- 在生產環境中，建議加入認證機制
- 考慮限制可以存取的檔案路徑範圍

## 擴展建議

### 1. 自動化映射表更新

可以建立一個腳本來自動掃描程式碼並更新映射表：

```bash
# 範例：掃描所有 handler 函數並生成映射表
go run scripts/generate_mappings.go
```

### 2. 整合 CI/CD

在 CI/CD pipeline 中：
- 檢查映射表是否與程式碼同步
- 自動更新行號
- 驗證所有映射的檔案都存在

### 3. 增強分析功能

- 加入歷史 duration 比較
- 提供效能趨勢分析
- 自動標記異常的 spans

### 4. LLM 整合

- 直接在 API 中整合 LLM 分析
- 提供一鍵分析功能
- 生成優化建議報告

## 範例場景

### 場景 1: 分析慢速 API

1. 使用者回報 `/api/order/create` 很慢
2. 在 Grafana 中找到一個慢速的 trace
3. 呼叫 source code API 獲取原始碼和 child spans
4. 發現 `processPayment` 子 span 佔用了 80% 的時間
5. 查看 `processPayment` 的原始碼，發現呼叫了外部 payment gateway
6. 優化：加入 timeout 和 retry 機制

### 場景 2: 新增自訂 handler

1. 建立新的 handler `handlers/analytics.go`
2. 新增映射：
```bash
curl -X POST http://localhost:8080/api/mappings \
  -H "Content-Type: application/json" \
  -d '{
    "mappings": [
      {
        "span_name": "GET /api/analytics",
        "file_path": "handlers/analytics.go",
        "function_name": "GetAnalytics",
        "start_line": 15,
        "end_line": 80,
        "description": "Analytics data retrieval"
      }
    ]
  }'
```
3. 測試 API 並驗證 tracing
4. 使用 source code API 確認映射正確

## 故障排除

### Tempo 連接問題

如果無法連接到 Tempo：

1. 檢查 Tempo 是否正在運行：
```bash
docker-compose ps tempo
```

2. 檢查 Tempo URL：
```bash
curl http://localhost:3200/api/search
```

3. 設定正確的環境變數：
```bash
export TEMPO_URL=http://localhost:3200
```

### 映射表問題

如果映射表沒有載入：

1. 檢查檔案是否存在：
```bash
ls -la source_code_mappings.json
```

2. 驗證 JSON 格式：
```bash
cat source_code_mappings.json | jq .
```

3. 手動重新載入：
```bash
curl -X POST http://localhost:8080/api/mappings/reload
```

### 原始碼讀取問題

如果無法讀取原始碼：

1. 檢查檔案路徑是否正確
2. 確認檔案權限
3. 驗證行號範圍是否有效

## 使用範例

### 完整流程：分析慢速 API

#### 1. 產生測試訂單

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

#### 2. 在 Grafana 找到 Trace

1. 訪問 http://localhost:3000
2. 進入 Explore → Tempo
3. 搜尋 `service.name="trace-demo-service"`
4. 找到 operation name 為 `POST /api/order/create` 的 trace
5. 複製 **Trace ID** 和 **Span ID**

#### 3. 獲取原始碼和分析資料

```bash
curl "http://localhost:8080/api/source-code?span_id=YOUR_SPAN_ID&trace_id=YOUR_TRACE_ID" | jq .
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
  "source_code": "func CreateOrder(w http.ResponseWriter, r *http.Request) {...}",
  "child_spans": [
    {"span_name": "validateOrder", "duration": "52.3ms"},
    {"span_name": "processPayment", "duration": "850.2ms"}
  ]
}
```

#### 4. 使用 LLM 分析

將回應提供給 LLM：

```
這個 API 的 duration 是 1.2s，比預期長。以下是原始碼和子 span 資訊：
[貼上 JSON]
請分析可能導致 duration 較長的原因，並提供優化建議。
```

### 與 Tempo Latency Anomaly Service 整合

```bash
# 1. 從 Anomaly Service 獲取最慢的 span
curl http://localhost:9090/v1/traces/{traceId}/longest-span

# 2. 使用 span name 查詢原始碼
curl -X POST http://localhost:8080/api/source-code \
  -H "Content-Type: application/json" \
  -d '{"spanName": "POST /api/order/create"}'
```

## 總結

這個 Source Code Analysis API 提供了一個強大的工具，讓您可以：

- 快速定位效能瓶頸
- 自動化效能分析流程
- 整合 LLM 進行智能分析
- 維護程式碼與 tracing 的對應關係

透過這個 API，您可以更有效地分析和優化應用程式的效能。
