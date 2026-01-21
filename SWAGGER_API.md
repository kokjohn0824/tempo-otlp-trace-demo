# Swagger API 文檔

本專案已整合 Swagger UI，提供互動式 API 文檔和測試介面。

## 功能特色

1. **互動式 API 文檔** - 透過 Swagger UI 瀏覽所有可用的 API
2. **即時測試** - 直接在瀏覽器中測試 API 端點
3. **自動生成** - 從程式碼註解自動生成 API 文檔
4. **完整的 Schema 定義** - 包含請求和回應的資料結構

## 快速開始

### 1. 生成 Swagger 文檔

首先需要安裝 `swag` CLI 工具並生成文檔：

```bash
# 安裝 swag 工具並生成文檔
make swagger

# 或手動執行
go install github.com/swaggo/swag/cmd/swag@latest
swag init -g main.go --output ./docs
```

### 2. 啟動服務

```bash
# 使用 Docker Compose 啟動所有服務
make up

# 或只啟動應用程式
make run
```

### 3. 訪問 Swagger UI

在瀏覽器中開啟：

```
http://localhost:8080/swagger/
```

或使用 Makefile 指令：

```bash
make open-swagger
```

## 可用的 API 端點

### Source Code APIs

#### 1. GET /api/span-names

獲取所有可用的 span names 列表。

**回應範例：**
```json
{
  "span_names": [
    {
      "span_name": "POST /api/order/create",
      "file_path": "handlers/order.go",
      "function_name": "CreateOrder",
      "description": "Handles order creation with comprehensive tracing",
      "start_line": 21,
      "end_line": 85
    }
  ],
  "count": 42
}
```

**使用場景：**
- 查看所有可追蹤的 span 操作
- 獲取 span 與原始碼的對應關係
- 用於 UI 下拉選單或自動完成功能

#### 2. POST /api/source-code

根據 span name 獲取對應的原始碼。

**請求範例：**
```json
{
  "spanName": "POST /api/order/create"
}
```

**回應範例：**
```json
{
  "span_name": "POST /api/order/create",
  "file_path": "handlers/order.go",
  "function_name": "CreateOrder",
  "start_line": 21,
  "end_line": 85,
  "source_code": "func CreateOrder(w http.ResponseWriter, r *http.Request) {\n\t// ... code ...\n}"
}
```

### Mapping Management APIs

#### 3. GET /api/mappings

獲取所有原始碼對應關係。

#### 4. POST /api/mappings

更新或新增原始碼對應關係。

#### 5. DELETE /api/mappings

刪除指定的原始碼對應關係。

**參數：**
- `span_name` (query parameter) - 要刪除的 span name

#### 6. POST /api/mappings/reload

從設定檔重新載入對應關係。

## 測試 API

### 使用 Swagger UI 測試

1. 訪問 http://localhost:8080/swagger/
2. 點擊任何 API 端點
3. 點擊 "Try it out" 按鈕
4. 填寫必要參數
5. 點擊 "Execute" 執行請求
6. 查看回應結果

### 使用測試腳本

我們提供了自動化測試腳本：

```bash
# 測試 span-names API
chmod +x scripts/test-span-names.sh
./scripts/test-span-names.sh

# 測試所有 API
make test-apis
```

### 使用 curl 測試

```bash
# 獲取所有 span names
curl http://localhost:8080/api/span-names | jq '.'

# 查詢特定 span 的原始碼
curl -X POST http://localhost:8080/api/source-code \
  -H "Content-Type: application/json" \
  -d '{"spanName": "POST /api/order/create"}' | jq '.'
```

## 整合到工作流程

### 與 Tempo Latency Anomaly Service 整合

完整的效能分析流程：

1. **從 Anomaly Service 獲取最慢的 span**
   ```bash
   curl http://localhost:9090/v1/traces/{traceId}/longest-span
   ```

2. **使用 span name 查詢可用的 spans**
   ```bash
   curl http://localhost:8080/api/span-names | jq '.span_names[] | select(.span_name == "...")'
   ```

3. **獲取原始碼進行分析**
   ```bash
   curl -X POST http://localhost:8080/api/source-code \
     -H "Content-Type: application/json" \
     -d '{"spanName": "POST /api/order/create"}'
   ```

### 與 LLM 整合

將 API 回應直接傳送給 LLM 進行程式碼分析：

```bash
# 獲取原始碼並傳送給 LLM
SOURCE_CODE=$(curl -X POST http://localhost:8080/api/source-code \
  -H "Content-Type: application/json" \
  -d '{"spanName": "POST /api/order/create"}' | jq -r '.source_code')

# 使用 LLM 分析
echo "請分析以下程式碼的效能瓶頸：\n${SOURCE_CODE}" | llm
```

## 開發指南

### 新增 API 端點的 Swagger 註解

在 handler 函數上方新增註解：

```go
// GetSourceCode handles requests to retrieve source code for a span
// @Summary Get source code for a span
// @Description Retrieves the source code associated with a specific span name
// @Tags Source Code
// @Accept json
// @Produce json
// @Param request body SourceCodeRequest true "Span name to query"
// @Success 200 {object} models.SourceCodeResponse
// @Failure 400 {string} string "Invalid request"
// @Failure 404 {string} string "Mapping not found"
// @Router /api/source-code [post]
func GetSourceCode(w http.ResponseWriter, r *http.Request) {
    // ... implementation ...
}
```

### 更新 Swagger 文檔

每次修改 API 註解後，需要重新生成文檔：

```bash
make swagger
```

### 在 Model 中新增範例

在 struct tag 中新增 `example` 標籤：

```go
type SourceCodeRequest struct {
    SpanName string `json:"spanName" example:"POST /api/order/create"`
}
```

## 常見問題

### Q: Swagger UI 顯示空白？

A: 確保已執行 `make swagger` 生成文檔，並且 `docs/` 目錄存在。

### Q: 如何自訂 Swagger 設定？

A: 修改 `main.go` 中的 Swagger 註解：

```go
// @title Tempo OTLP Trace Demo API
// @version 1.0
// @description API for generating traces and retrieving source code mappings
// @host localhost:8080
// @BasePath /
```

### Q: 如何在生產環境中使用？

A: 建議在生產環境中：
1. 使用環境變數控制 Swagger UI 的啟用/停用
2. 設定適當的 CORS 策略
3. 新增認證機制保護 API

## 相關資源

- [Swagger 官方文檔](https://swagger.io/docs/)
- [swaggo/swag GitHub](https://github.com/swaggo/swag)
- [OpenAPI 規範](https://spec.openapis.org/oas/latest.html)

## 下一步

- 探索 [整合測試文檔](../test-integration.sh)
- 查看 [架構改進建議](../ARCHITECTURE_IMPROVEMENT.md)
- 閱讀 [API 使用範例](USAGE_EXAMPLE.md)
