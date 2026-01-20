# Tempo OTLP Trace Demo - 實作計劃

## 專案目標

建立一個最小化的 trace 發送系統，確保真實的 traces 能夠正確地送到 Tempo，並且包含正確的 parent/child 關係和 durations，以支援後續的「最長 span 分析」邏輯。

## 架構選擇

### 推薦架構（最小化堆疊）

```
Service(s) → OTLP (gRPC/HTTP) → OpenTelemetry Collector → Tempo → Grafana (可選)
```

**選擇理由：**
- 保持應用程式配置簡單
- Collector 可以處理 batching、retries、filtering
- 未來可以輕鬆擴展而不需要修改程式碼

### 替代架構（更簡單但功能較少）

```
App → OTLP exporter → Tempo OTLP receiver (4317 gRPC 或 4318 HTTP)
```

## 關鍵注意事項（常見錯誤）

### 1. Tempo OTLP 綁定地址
- **問題：** Tempo 預設 OTLP receivers 綁定到 localhost:4317/4318，在 Docker 容器間會無法接收
- **解決：** 設定為 `0.0.0.0`

### 2. 採樣率設定
- **問題：** 採樣可能會丟棄「異常但罕見」的長 traces
- **解決：** Demo 環境強制 100% 採樣 (parent-based 1.0)

### 3. Context Propagation
- **問題：** 服務間如果沒有正確傳遞 W3C traceparent，會產生多個 traceID
- **解決：** 確保每個服務都正確 extract/inject traceparent（OTel auto-instrumentation 通常會自動處理）

## 技術選擇

### 語言/框架選擇

#### 選項 A：Go (推薦用於此 Demo)
**優點：**
- 不需要 agent JAR
- 程式碼簡潔、可預測
- 適合多服務 trace propagation demo
- 容易建立獨立的測試 trace generator

**缺點：**
- 需要手動寫 spans/instrumentation
- 不會反映真實 Java/WildFly stack 的行為

#### 選項 B：Java + OpenTelemetry Agent
**優點：**
- 零程式碼變更（auto-instrumentation）
- 自動處理 Spring MVC、RestTemplate、JDBC 等
- 最接近真實生產環境

**缺點：**
- 需要下載並配置 agent JAR
- JVM 參數配置較複雜
- WildFly 可能有 classloader 問題

## 實作步驟

### Phase 1: 基礎設施設定

#### 1.1 Docker Compose 配置
建立 `docker-compose.yml` 包含：
- OpenTelemetry Collector
- Tempo
- Grafana (可選，用於視覺化)

#### 1.2 OpenTelemetry Collector 配置
建立 `otel-collector.yaml`：
```yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318

processors:
  batch: {}

exporters:
  otlp:
    endpoint: tempo:4317
    tls:
      insecure: true

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlp]
```

#### 1.3 Tempo 配置
建立 `tempo.yaml`：
```yaml
distributor:
  receivers:
    otlp:
      protocols:
        grpc:
          endpoint: 0.0.0.0:4317
        http:
          endpoint: 0.0.0.0:4318
```

### Phase 2: 應用程式開發（Go 版本）

#### 2.1 專案結構
```
tempo-otlp-trace-demo/
├── go.mod
├── go.sum
├── main.go
├── handlers/
│   ├── order.go
│   ├── user.go
│   ├── report.go
│   ├── search.go
│   ├── batch.go
│   └── simulate.go
├── tracing/
│   └── helpers.go
├── models/
│   └── request.go
├── docker-compose.yml
├── otel-collector.yaml
├── tempo.yaml
├── scripts/
│   └── test-apis.sh
├── README.md
└── PLAN.md
```

#### 2.2 Go 應用程式功能
建立 `main.go` 包含：
- HTTP server (port 8080)
- OTLP gRPC exporter 配置
- Tracer provider 初始化
- 100% 採樣率設定
- 多個 API endpoints 模擬真實世界場景

#### 2.3 API Endpoints 設計（模擬真實世界場景）

##### API 1: `/api/order/create` - 電商訂單建立
模擬一個完整的訂單建立流程，包含：
- **Root Span**: `POST /api/order/create` (HTTP handler)
- **Child Spans**:
  - `validateOrder` - 驗證訂單資料 (50-100ms)
  - `checkInventory` - 檢查庫存 (100-200ms)
  - `calculatePrice` - 計算價格 (30-80ms)
  - `processPayment` - 處理付款 (200-500ms)
    - Nested: `callPaymentGateway` (150-400ms)
    - Nested: `recordTransaction` (20-50ms)
  - `createShipment` - 建立出貨單 (80-150ms)
  - `sendNotification` - 發送通知 (50-100ms)
    - Nested: `sendEmail` (30-60ms)
    - Nested: `sendSMS` (20-40ms)
  - `saveToDatabase` - 儲存到資料庫 (100-300ms)

**預期總時長**: 600-1500ms
**Span 數量**: 10-12 個

##### API 2: `/api/user/profile` - 使用者資料查詢
模擬簡單的查詢操作：
- **Root Span**: `GET /api/user/profile`
- **Child Spans**:
  - `authenticate` - 驗證身份 (20-50ms)
  - `queryDatabase` - 查詢資料庫 (50-150ms)
  - `loadPreferences` - 載入偏好設定 (30-80ms)
  - `formatResponse` - 格式化回應 (10-30ms)

**預期總時長**: 110-310ms
**Span 數量**: 4-5 個

##### API 3: `/api/report/generate` - 報表生成（異常長）
模擬需要較長時間的報表生成：
- **Root Span**: `POST /api/report/generate`
- **Child Spans**:
  - `validateRequest` - 驗證請求 (30-60ms)
  - `fetchDataFromMultipleSources` - 從多個來源取得資料 (500-1000ms)
    - Nested: `queryMainDB` (200-400ms)
    - Nested: `queryAnalyticsDB` (200-400ms)
    - Nested: `fetchExternalAPI` (100-300ms)
  - `processData` - 處理資料 (300-800ms)
    - Nested: `aggregateData` (150-400ms)
    - Nested: `calculateMetrics` (150-400ms)
  - `generatePDF` - 生成 PDF (400-1200ms)
  - `uploadToStorage` - 上傳到儲存 (200-500ms)
  - `notifyUser` - 通知使用者 (50-100ms)

**預期總時長**: 1500-3500ms（這是「異常長」的 trace）
**Span 數量**: 10-12 個

##### API 4: `/api/search` - 搜尋功能
模擬搜尋引擎查詢：
- **Root Span**: `GET /api/search`
- **Child Spans**:
  - `parseQuery` - 解析查詢 (10-30ms)
  - `searchIndex` - 搜尋索引 (80-200ms)
  - `rankResults` - 排序結果 (40-100ms)
  - `fetchDetails` - 取得詳細資訊 (60-150ms)
    - Nested: `batchQuery` (50-130ms)
  - `applyFilters` - 套用篩選 (20-50ms)

**預期總時長**: 210-530ms
**Span 數量**: 6-7 個

##### API 5: `/api/batch/process` - 批次處理
模擬批次處理多個項目：
- **Root Span**: `POST /api/batch/process`
- **Child Spans**:
  - `validateBatch` - 驗證批次 (30-60ms)
  - `processItems` - 處理項目 (動態，依據項目數量)
    - Nested: `processItem-1` (50-150ms)
    - Nested: `processItem-2` (50-150ms)
    - Nested: `processItem-3` (50-150ms)
    - ... (可配置 3-10 個項目)
  - `aggregateResults` - 彙總結果 (40-80ms)
  - `saveResults` - 儲存結果 (100-200ms)

**預期總時長**: 300-1500ms（依項目數量而定）
**Span 數量**: 6-15 個

##### API 6: `/api/simulate` - 自訂模擬（測試用）
可透過參數自訂 trace 特性：
- Query parameters:
  - `depth`: span 巢狀深度 (1-10)
  - `breadth`: 每層的 span 數量 (1-5)
  - `duration`: 每個 span 的平均時長 (ms)
  - `variance`: 時長變異度 (0.0-1.0)

**範例**: `/api/simulate?depth=5&breadth=3&duration=100&variance=0.5`

#### 2.4 Span 屬性設定
每個 span 應包含真實的屬性：
- `http.method`: GET, POST 等
- `http.route`: API 路徑
- `http.status_code`: 200, 400, 500 等
- `db.system`: postgresql, mysql 等（模擬）
- `db.statement`: SQL 查詢（模擬）
- `error`: true/false（可選，模擬錯誤情況）
- 自訂屬性：`operation.type`, `item.count`, `user.id` 等

#### 2.5 環境變數
```bash
OTEL_EXPORTER_OTLP_ENDPOINT=otel-collector:4317  # 在 Docker 內
OTEL_SERVICE_NAME=trace-demo-service
```

#### 2.6 程式碼結構建議
```
tempo-otlp-trace-demo/
├── main.go              # 主程式、tracer 初始化
├── handlers/
│   ├── order.go         # 訂單相關 API
│   ├── user.go          # 使用者相關 API
│   ├── report.go        # 報表相關 API
│   ├── search.go        # 搜尋相關 API
│   ├── batch.go         # 批次處理 API
│   └── simulate.go      # 自訂模擬 API
├── tracing/
│   └── helpers.go       # Tracing 輔助函數
└── models/
    └── request.go       # 請求/回應模型
```

### Phase 3: 驗證與測試

#### 3.1 Ingestion 驗證
- [ ] Collector logs 顯示 spans 被接收和導出
- [ ] Tempo 可以通過 traceID 查詢到 trace
- [ ] 可選：加入 debug/logging exporter 到 Collector

#### 3.2 Trace 品質驗證
- [ ] 每個請求至少有一個 SERVER span
- [ ] 存在巢狀 spans（child spans）
- [ ] Durations 看起來合理（不是全部 ~0ms）
- [ ] Parent/child 關係正確
- [ ] 各 API 產生的 span 數量符合預期
- [ ] 異常長的 trace (`/api/report/generate`) 可以被正確識別

#### 3.3 屬性驗證
- [ ] `service.name` 正確設定
- [ ] Resource attributes 存在
- [ ] Span attributes 正確（如 `http.route`）

### Phase 4: 擴展功能（可選）

#### 4.1 多服務 Demo
- 建立第二個服務
- 實作服務間的 HTTP 呼叫
- 驗證 trace propagation（單一 traceID 跨服務）

#### 4.2 Grafana 整合
- 配置 Tempo datasource
- 建立基本的 trace 查詢面板
- 測試「最長 span」查詢邏輯

#### 4.3 測試場景腳本
建立測試腳本來產生各種 trace patterns：
- **正常流量**: 混合呼叫各個 API
- **壓力測試**: 大量並發請求
- **異常檢測**: 定期產生異常長的 traces
- **錯誤情境**: 模擬失敗的請求（4xx, 5xx）

範例測試腳本：
```bash
# 正常流量
for i in {1..10}; do
  curl http://localhost:8080/api/user/profile
  curl http://localhost:8080/api/search?q=test
  sleep 1
done

# 產生異常長的 trace
curl -X POST http://localhost:8080/api/report/generate

# 批次處理測試
curl -X POST http://localhost:8080/api/batch/process -d '{"items": 10}'
```

## 配置參考

### Go OTLP Exporter 關鍵配置

```go
// gRPC exporter
exp, err := otlptracegrpc.New(ctx,
    otlptracegrpc.WithEndpoint("otel-collector:4317"),
    otlptracegrpc.WithInsecure(),
)

// 100% 採樣
tp := sdktrace.NewTracerProvider(
    sdktrace.WithResource(res),
    sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(1.0))),
    sdktrace.WithBatcher(exp),
)
```

### Java Agent 配置（如果選擇 Java）

```bash
java \
  -javaagent:/path/opentelemetry-javaagent.jar \
  -Dotel.service.name=demo-svc \
  -Dotel.traces.exporter=otlp \
  -Dotel.metrics.exporter=none \
  -Dotel.logs.exporter=none \
  -Dotel.exporter.otlp.endpoint=http://otel-collector:4318 \
  -Dotel.exporter.otlp.protocol=http/protobuf \
  -Dotel.traces.sampler=parentbased_traceidratio \
  -Dotel.traces.sampler.arg=1.0 \
  -jar app.jar
```

## 常見問題排查

### 問題：看不到 traces

**檢查清單：**
1. Tempo OTLP receiver 是否綁定到 `0.0.0.0`（不是 `localhost`）
2. Endpoint 是否可達：
   - Docker 內：使用 service name（如 `tempo:4317`）
   - Host 上：使用 `127.0.0.1:4317`（確保 port 有 publish）
3. 採樣率是否設為 1.0
4. 檢查 Collector logs 是否有錯誤

### 問題：Durations 全部是 0ms

**可能原因：**
- Container/VM 時鐘問題
- 手動 instrumentation 的 timestamp 不正確
- 使用 auto-instrumentation 通常可避免此問題

### 問題：多個 traceID（應該只有一個）

**可能原因：**
- Context propagation 中斷
- 服務間沒有正確傳遞 W3C traceparent header
- 檢查 propagator 設定

## 成功標準

1. ✅ 能夠產生並發送 traces 到 Tempo
2. ✅ Traces 包含正確的 parent/child 關係
3. ✅ Span durations 準確且有意義
4. ✅ 可以通過 traceID 在 Tempo 中查詢到完整 trace
5. ✅ 支援產生不同 duration 的 traces 用於測試「最長 span」邏輯
6. ✅ 程式碼簡潔、易於理解和修改
7. ✅ 至少實作 5 個不同的 API endpoints，模擬真實世界場景
8. ✅ 每個 API 產生 4-15 個 spans，包含巢狀結構
9. ✅ Span 包含有意義的屬性（http.method, db.system 等）
10. ✅ 可以明確識別出「異常長」的 trace（如報表生成 API）

## 未來擴展方向

1. 整合到 CI/CD pipeline 作為 trace generator
2. 建立自動化測試來驗證「最長 span」分析邏輯
3. 支援更多的 trace patterns（錯誤、重試、timeout 等）
4. 加入 metrics 和 logs correlation
5. 效能測試：大量 traces 的產生和處理

## 參考資源

- OpenTelemetry Collector 官方文件
- Grafana Tempo 官方文件
- OpenTelemetry Go SDK 文件
- W3C Trace Context 規範
