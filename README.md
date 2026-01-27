# Tempo OTLP Trace Demo

[![GitHub](https://img.shields.io/badge/GitHub-tempo--otlp--trace--demo-blue?logo=github)](https://github.com/kokjohn0824/tempo-otlp-trace-demo)
[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go)](https://go.dev/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?logo=docker)](https://www.docker.com/)
[![Swagger](https://img.shields.io/badge/API-Swagger-85EA2D?logo=swagger)](http://localhost:8080/swagger/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

一個完整的 OpenTelemetry 追蹤示範專案，用於產生真實世界的 trace 資料並發送到 Grafana Tempo。

## 📚 文件導覽

### 核心文件
- **[README.md](README.md)** - 專案說明和使用指南（本文件）
- **[INSTALLATION.md](INSTALLATION.md)** - 詳細的安裝和設定指南
- **[CONTRIBUTING.md](CONTRIBUTING.md)** - 貢獻指南
- **[CHANGELOG.md](CHANGELOG.md)** - 變更日誌

### Makefile 相關
- **[QUICK_REFERENCE.md](QUICK_REFERENCE.md)** - Makefile 快速參考
- **[MAKEFILE_GUIDE.md](MAKEFILE_GUIDE.md)** - Makefile 詳細使用指南

### 原始碼分析功能
- **[SOURCE_CODE_API.md](SOURCE_CODE_API.md)** - 原始碼分析 API 文件
- **[USAGE_EXAMPLE.md](USAGE_EXAMPLE.md)** - 原始碼分析 API 使用範例

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
**預期時長**: 600-1500ms (正常) / 5600-6500ms (sleep=true)  
**Span 數量**: 10-12 個  
**說明**: 模擬電商訂單建立流程，包含驗證、庫存檢查、付款處理、出貨和通知

**參數說明**:
| 參數 | 類型 | 必填 | 說明 |
|------|------|------|------|
| user_id | string | 是 | 使用者 ID |
| product_id | string | 是 | 產品 ID |
| quantity | int | 是 | 購買數量 |
| price | float | 是 | 單價 |
| sleep | bool | 否 | 若為 true，則在 processPayment 子操作中額外等待 5 秒，用於模擬異常延遲 |

**範例請求 (正常)**:
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

**範例請求 (模擬異常延遲)**:
```bash
curl -X POST http://localhost:8080/api/order/create \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user_12345",
    "product_id": "prod_98765",
    "quantity": 2,
    "price": 299.99,
    "sleep": true
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

## 🆕 原始碼分析 API

這個專案現在包含了強大的原始碼分析功能，可以根據 Tempo 中的 span 資訊來獲取對應的原始碼，以供 LLM 分析效能問題。

### 主要功能

1. **自動原始碼映射**: 根據 span name 自動找到對應的原始碼位置
2. **完整的 span 資訊**: 包含 duration、attributes、child spans 等
3. **LLM 友善的輸出**: JSON 格式，可直接提供給 LLM 分析
4. **映射表管理**: 支援新增、更新、刪除和重新載入映射

### 新增的 API Endpoints

#### 1. 獲取原始碼
```bash
GET /api/source-code?span_id={spanId}&trace_id={traceId}
```
根據 span ID 和 trace ID 獲取對應的原始碼及相關資訊。

#### 2. 管理映射表
```bash
GET /api/mappings              # 查詢所有映射
POST /api/mappings             # 新增/更新映射
DELETE /api/mappings?span_name={name}  # 刪除映射
POST /api/mappings/reload      # 重新載入映射
```

### 快速使用範例

```bash
# 1. 產生一個 trace
curl -X POST http://localhost:8080/api/order/create \
  -H "Content-Type: application/json" \
  -d '{"user_id":"user_123","product_id":"prod_456","quantity":2,"price":99.99}'

# 2. 在 Grafana 中找到 trace ID 和 span ID

# 3. 獲取原始碼和分析資料
curl "http://localhost:8080/api/source-code?span_id=YOUR_SPAN_ID&trace_id=YOUR_TRACE_ID" | jq .

# 4. 將結果提供給 LLM 分析效能瓶頸
```

### 詳細文件

- **[SOURCE_CODE_API.md](SOURCE_CODE_API.md)** - 完整的 API 文件和參考
- **[USAGE_EXAMPLE.md](USAGE_EXAMPLE.md)** - 詳細的使用範例和場景

### 測試原始碼分析 API

```bash
# 執行測試腳本
./scripts/test-source-code-api.sh
```

## 快速開始

### 前置需求

- Docker 和 Docker Compose
- Go 1.21+ (如果要在本地執行)
- curl 和 jq (用於測試腳本)
- Make (用於執行 Makefile 指令)

### 使用 Makefile (推薦)

本專案提供了完整的 Makefile 來簡化開發、測試和部署流程。

#### 查看所有可用指令

```bash
make help
```

#### 常用指令

**啟動所有服務**:
```bash
make up
```

**檢查服務健康狀態**:
```bash
make health
```

**執行 API 測試**:
```bash
make test-apis
```

**查看日誌**:
```bash
make logs          # 所有服務
make logs-app      # 應用程式
make logs-collector # OTel Collector
make logs-tempo    # Tempo
make logs-grafana  # Grafana
```

**停止服務**:
```bash
make down
```

**開發模式** (啟動基礎設施並在本地運行應用程式):
```bash
make dev
```

**建立和部署**:
```bash
make build              # 編譯應用程式
make docker-build       # 建立 Docker 映像
make deploy            # 建立並推送映像
```

**清理**:
```bash
make clean             # 清理編譯產物
make clean-all         # 完全清理 (包含 Docker)
```

### 手動啟動 (不使用 Makefile)

如果你偏好手動操作：

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

**使用 Makefile (推薦)**:
```bash
make test-apis        # 完整測試
make test-quick       # 快速測試 (減少等待時間)
```

**手動執行測試腳本**:
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

**使用 Makefile (推薦)**:
```bash
make dev
```

這個指令會自動：
1. 啟動基礎設施 (OTel Collector, Tempo, Grafana)
2. 編譯並執行應用程式
3. 檢查服務健康狀態

**手動執行**:

1. **啟動基礎設施**（不含應用程式）:
```bash
docker-compose up -d otel-collector tempo-server grafana
# 或使用 Makefile
make infra-up
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
# 或使用 Makefile
make run
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
└── README.md             # 本文件
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

**使用 Makefile**:
```bash
make down              # 停止所有服務
make down-volumes      # 停止並刪除 volumes（會提示確認）
```

**手動執行**:
```bash
# 停止所有服務
docker-compose down

# 停止並刪除 volumes（清除所有資料）
docker-compose down -v
```

## Makefile 指令參考

### 開發相關

| 指令 | 說明 |
|------|------|
| `make help` | 顯示所有可用指令 |
| `make check-deps` | 檢查必要的依賴工具 |
| `make install-deps` | 安裝 Go 依賴 |
| `make fmt` | 格式化 Go 程式碼 |
| `make vet` | 執行 Go vet 檢查 |
| `make lint` | 執行 golangci-lint 檢查 |
| `make build` | 編譯 Go 應用程式 (Linux) |
| `make build-local` | 編譯本地版本 |
| `make run` | 在本地執行應用程式 |
| `make dev` | 開發模式 (啟動基礎設施並運行應用) |

### Docker 相關

| 指令 | 說明 |
|------|------|
| `make docker-build` | 建立 Docker 映像 |
| `make docker-push` | 推送映像到 Registry |
| `make up` | 啟動所有服務 |
| `make down` | 停止所有服務 |
| `make down-volumes` | 停止服務並刪除 volumes |
| `make restart` | 重啟所有服務 |
| `make infra-up` | 只啟動基礎設施 |

### 日誌和監控

| 指令 | 說明 |
|------|------|
| `make logs` | 查看所有服務日誌 |
| `make logs-app` | 查看應用程式日誌 |
| `make logs-collector` | 查看 OTel Collector 日誌 |
| `make logs-tempo` | 查看 Tempo 日誌 |
| `make logs-grafana` | 查看 Grafana 日誌 |
| `make ps` | 查看服務狀態 |
| `make health` | 檢查所有服務健康狀態 |

### 測試相關

| 指令 | 說明 |
|------|------|
| `make test` | 執行 Go 單元測試 |
| `make test-coverage` | 執行測試並生成覆蓋率報告 |
| `make test-apis` | 執行 API 測試腳本 |
| `make test-quick` | 快速 API 測試 |
| `make bench` | 執行效能測試 |

### 清理和維護

| 指令 | 說明 |
|------|------|
| `make clean` | 清理編譯產物 |
| `make clean-all` | 完全清理 (包含 Docker) |
| `make tidy` | 整理 Go 依賴 |

### CI/CD

| 指令 | 說明 |
|------|------|
| `make ci` | CI 流程 (格式化、檢查、測試、建立) |
| `make deploy` | 部署 (建立並推送映像) |
| `make all` | 完整建立流程 |

### 其他

| 指令 | 說明 |
|------|------|
| `make open-grafana` | 在瀏覽器開啟 Grafana |
| `make open-app` | 在瀏覽器開啟應用程式 |

### 環境變數

Makefile 支援以下環境變數：

- `DOCKER_REGISTRY`: Docker Registry 位址 (用於推送映像)
- `DOCKER_TAG`: Docker 映像標籤 (預設: `latest`)
- `BASE_URL`: API 測試的基礎 URL (預設: `http://localhost:8080`)
- `PORT`: 應用程式 port (預設: `8080`)

**範例**:
```bash
# 建立並推送映像到自訂 Registry
make deploy DOCKER_REGISTRY=myregistry.com DOCKER_TAG=v1.0.0

# 使用自訂 port 執行應用程式
make run PORT=9090

# 測試遠端服務
make test-apis BASE_URL=http://production-server:8080
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
