# Tempo OTLP Trace Demo

一個完整的 OpenTelemetry 追蹤示範專案，用於產生真實世界的 trace 資料並發送到 Grafana Tempo。

## 專案目標

建立一個最小化的 trace 發送系統，確保真實的 traces 能夠正確地送到 Tempo，並且包含正確的 parent/child 關係和 durations，以支援後續的「最長 span 分析」邏輯。

## 架構

```
Go Application → OTLP (gRPC) → OpenTelemetry Collector → Tempo → Grafana
```

### 元件說明

- **Go Application**: 提供多個 API endpoints，每個模擬不同的真實世界場景
- **OpenTelemetry Collector**: 接收 traces、批次處理、並轉發到 Tempo
- **Grafana Tempo**: 儲存和查詢 traces
- **Grafana**: 視覺化介面，用於瀏覽和分析 traces

## API Endpoints

所有 API 都會產生具有真實 parent/child 關係的 traces：

### 1. `/api/order/create` - 訂單建立
**方法**: POST  
**預期時長**: 600-1500ms  
**Span 數量**: 10-12 個  
**說明**: 模擬電商訂單建立流程，包含驗證、庫存檢查、付款處理、出貨和通知

**範例請求**:
```bash
curl -X POST http://localhost:8080/api/order/create \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user_12345",
    "product_id": "prod_98765",
    "quantity": 2,
    "price": 299.99
  }'
```

### 2. `/api/user/profile` - 使用者資料查詢
**方法**: GET  
**預期時長**: 110-310ms  
**Span 數量**: 4-5 個  
**說明**: 簡單的查詢操作，包含認證、資料庫查詢和偏好設定載入

**範例請求**:
```bash
curl http://localhost:8080/api/user/profile?user_id=user_12345
```

### 3. `/api/report/generate` - 報表生成 ⭐ **異常長的 Trace**
**方法**: POST  
**預期時長**: 1500-3500ms  
**Span 數量**: 10-12 個  
**說明**: 模擬需要較長時間的報表生成，包含多資料源查詢、資料處理和 PDF 生成

**範例請求**:
```bash
curl -X POST http://localhost:8080/api/report/generate \
  -H "Content-Type: application/json" \
  -d '{
    "report_type": "sales",
    "start_date": "2024-01-01",
    "end_date": "2024-01-31",
    "filters": ["region:US", "category:electronics"]
  }'
```

### 4. `/api/search` - 搜尋功能
**方法**: GET  
**預期時長**: 210-530ms  
**Span 數量**: 6-7 個  
**說明**: 模擬搜尋引擎查詢，包含查詢解析、索引搜尋和結果排序

**範例請求**:
```bash
curl "http://localhost:8080/api/search?q=laptop&page=1&limit=10"
```

### 5. `/api/batch/process` - 批次處理
**方法**: POST  
**預期時長**: 300-1500ms（依項目數量而定）  
**Span 數量**: 6-15 個  
**說明**: 批次處理多個項目，每個項目有獨立的 span

**範例請求**:
```bash
curl -X POST http://localhost:8080/api/batch/process \
  -H "Content-Type: application/json" \
  -d '{
    "items": ["item1", "item2", "item3", "item4", "item5"]
  }'
```

### 6. `/api/simulate` - 自訂模擬
**方法**: GET  
**預期時長**: 可配置  
**Span 數量**: 可配置  
**說明**: 透過參數自訂 trace 特性，用於測試不同的 trace patterns

**參數**:
- `depth`: span 巢狀深度 (1-10)
- `breadth`: 每層的 span 數量 (1-5)
- `duration`: 每個 span 的平均時長 (ms)
- `variance`: 時長變異度 (0.0-1.0)

**範例請求**:
```bash
curl "http://localhost:8080/api/simulate?depth=5&breadth=3&duration=100&variance=0.5"
```

## 快速開始

### 前置需求

- Docker 和 Docker Compose
- Go 1.21+ (如果要在本地執行)
- curl 和 jq (用於測試腳本)

### 啟動服務

1. **啟動所有服務**:
```bash
docker-compose up -d
```

2. **查看日誌**:
```bash
# 查看所有服務
docker-compose logs -f

# 查看特定服務
docker-compose logs -f trace-demo-app
docker-compose logs -f otel-collector
docker-compose logs -f tempo
```

3. **檢查服務狀態**:
```bash
# 應用程式健康檢查
curl http://localhost:8080/health

# OTel Collector 健康檢查
curl http://localhost:13133/

# Tempo 健康檢查
curl http://localhost:3200/ready
```

### 執行測試

使用提供的測試腳本來產生各種 traces：

```bash
./scripts/test-apis.sh
```

這個腳本會：
- 測試所有 API endpoints
- 產生不同長度和複雜度的 traces
- 產生多個「異常長」的 traces 用於分析

### 查看 Traces

1. 開啟 Grafana: http://localhost:3000
2. 前往 **Explore** (左側選單的指南針圖示)
3. 選擇 **Tempo** 資料源
4. 搜尋 traces:
   - 依 Service Name: `trace-demo-service`
   - 依 Operation: 選擇特定的 API endpoint
   - 依 Duration: 找出最長的 traces

### 本地開發模式

如果要在本地執行應用程式（不使用 Docker）：

1. **啟動基礎設施**（不含應用程式）:
```bash
docker-compose up -d otel-collector tempo grafana
```

2. **設定環境變數**:
```bash
export OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317
export OTEL_SERVICE_NAME=trace-demo-service
export PORT=8080
```

3. **執行應用程式**:
```bash
go run main.go
```

## 專案結構

```
tempo-otlp-trace-demo/
├── handlers/              # API handlers
│   ├── order.go          # 訂單相關 API
│   ├── user.go           # 使用者相關 API
│   ├── report.go         # 報表相關 API
│   ├── search.go         # 搜尋相關 API
│   ├── batch.go          # 批次處理 API
│   └── simulate.go       # 自訂模擬 API
├── tracing/              # Tracing 相關程式碼
│   └── helpers.go        # Tracer 初始化和輔助函數
├── models/               # 資料模型
│   └── request.go        # 請求/回應結構
├── scripts/              # 工具腳本
│   └── test-apis.sh      # API 測試腳本
├── main.go               # 主程式
├── docker-compose.yml    # Docker Compose 配置
├── Dockerfile            # 應用程式 Docker 映像
├── otel-collector.yaml   # OTel Collector 配置
├── tempo.yaml            # Tempo 配置
├── grafana-datasources.yaml  # Grafana 資料源配置
├── go.mod                # Go 模組定義
├── go.sum                # Go 依賴校驗
├── README.md             # 本文件
└── PLAN.md               # 實作計劃
```

## Span 屬性

每個 span 都包含有意義的屬性，模擬真實應用程式：

- **HTTP 相關**: `http.method`, `http.route`, `http.status_code`
- **資料庫相關**: `db.system`, `db.statement`, `db.table`
- **業務邏輯**: `user.id`, `order.id`, `operation.type`
- **錯誤處理**: `error`, `error.reason`

## 配置說明

### 環境變數

- `OTEL_EXPORTER_OTLP_ENDPOINT`: OTel Collector 的 endpoint (預設: `localhost:4317`)
- `OTEL_SERVICE_NAME`: 服務名稱 (預設: `trace-demo-service`)
- `PORT`: HTTP 伺服器 port (預設: `8080`)

### 採樣率

目前設定為 **100% 採樣** (`TraceIDRatioBased(1.0)`)，確保所有 traces 都被記錄。生產環境應該調整為適當的採樣率。

## 常見問題排查

### 問題：看不到 traces

**檢查清單**:
1. 確認所有服務都在運行: `docker-compose ps`
2. 檢查 OTel Collector logs: `docker-compose logs otel-collector`
3. 檢查 Tempo logs: `docker-compose logs tempo`
4. 確認 Tempo 可以接收資料: `curl http://localhost:3200/ready`

### 問題：Traces 沒有 parent/child 關係

**可能原因**:
- Context propagation 問題
- 檢查程式碼中是否正確傳遞 context

### 問題：Duration 看起來不正確

**可能原因**:
- 時鐘同步問題（容器間）
- 檢查 Docker 時間設定

## 使用案例

### 1. 測試「最長 span」分析邏輯

```bash
# 產生多個 traces
for i in {1..10}; do
  curl -X POST http://localhost:8080/api/report/generate \
    -H "Content-Type: application/json" \
    -d '{"report_type":"test","start_date":"2024-01-01","end_date":"2024-12-31"}'
  sleep 1
done

# 在 Grafana 中搜尋最長的 traces
# 應該會看到 /api/report/generate 的 traces 明顯比其他長
```

### 2. 壓力測試

```bash
# 使用 Apache Bench 或類似工具
ab -n 1000 -c 10 http://localhost:8080/api/user/profile?user_id=test
```

### 3. 自訂 Trace Patterns

```bash
# 產生深度巢狀的 traces
curl "http://localhost:8080/api/simulate?depth=10&breadth=1&duration=50&variance=0.2"

# 產生寬度較大的 traces
curl "http://localhost:8080/api/simulate?depth=2&breadth=5&duration=100&variance=0.5"
```

## 停止服務

```bash
# 停止所有服務
docker-compose down

# 停止並刪除 volumes（清除所有資料）
docker-compose down -v
```

## 貢獻

歡迎提出 issues 和 pull requests！

## 授權

MIT License

## 參考資源

- [OpenTelemetry 官方文件](https://opentelemetry.io/docs/)
- [Grafana Tempo 文件](https://grafana.com/docs/tempo/latest/)
- [OpenTelemetry Go SDK](https://github.com/open-telemetry/opentelemetry-go)
- [W3C Trace Context](https://www.w3.org/TR/trace-context/)
