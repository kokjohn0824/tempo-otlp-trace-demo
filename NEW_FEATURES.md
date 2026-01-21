# 新功能說明

本次更新為 Tempo OTLP Trace Demo 新增了兩個重要功能：

## 1. Span Names API

新增了 `/api/span-names` 端點，可以查詢所有可用的 span names 及其對應的原始碼資訊。

### 使用方式

```bash
# 獲取所有可用的 span names
curl http://localhost:8080/api/span-names | jq '.'
```

### 回應範例

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
  "count": 41
}
```

### 應用場景

- **自動完成功能**：在 UI 中提供 span name 的自動完成選項
- **文檔生成**：自動生成所有可追蹤操作的文檔
- **測試工具**：快速查看系統中所有可分析的操作
- **整合測試**：驗證所有 span 都有對應的原始碼映射

## 2. Swagger UI 整合

整合了 Swagger UI，提供互動式 API 文檔和測試介面。

### 訪問 Swagger UI

在瀏覽器中開啟：

```
http://localhost:8080/swagger/
```

或使用 Makefile 指令：

```bash
make open-swagger
```

### Swagger UI 功能

1. **互動式文檔**：瀏覽所有 API 端點的詳細說明
2. **即時測試**：直接在瀏覽器中測試 API
3. **Schema 定義**：查看請求和回應的資料結構
4. **範例資料**：每個 API 都包含範例請求和回應

### 可用的 API 分類

#### Source Code APIs
- `GET /api/span-names` - 獲取所有可用的 span names
- `POST /api/source-code` - 根據 span name 獲取原始碼

#### Mappings APIs
- `GET /api/mappings` - 獲取所有映射
- `POST /api/mappings` - 更新映射
- `DELETE /api/mappings` - 刪除映射
- `POST /api/mappings/reload` - 重新載入映射

## 快速開始

### 1. 啟動服務

```bash
# 使用 Docker Compose
make up

# 或手動啟動
docker-compose up -d
```

### 2. 測試新功能

```bash
# 測試 span-names API
./scripts/test-span-names.sh

# 或手動測試
curl http://localhost:8080/api/span-names | jq '.count'
```

### 3. 訪問 Swagger UI

```bash
# 在瀏覽器中開啟
make open-swagger

# 或直接訪問
open http://localhost:8080/swagger/
```

## 整合流程範例

完整的效能分析流程：

```bash
# 1. 從 Anomaly Service 獲取異常 trace
TRACE_ID="abc123"

# 2. 獲取 longest span
LONGEST_SPAN=$(curl -s "http://localhost:9090/v1/traces/${TRACE_ID}/longest-span")
SPAN_NAME=$(echo "$LONGEST_SPAN" | jq -r '.span.name')

# 3. 檢查 span name 是否有映射
curl -s http://localhost:8080/api/span-names | jq ".span_names[] | select(.span_name == \"$SPAN_NAME\")"

# 4. 獲取原始碼
curl -X POST http://localhost:8080/api/source-code \
  -H "Content-Type: application/json" \
  -d "{\"spanName\": \"$SPAN_NAME\"}" | jq '.source_code'
```

## 開發指南

### 更新 Swagger 文檔

每次修改 API 後，需要重新生成 Swagger 文檔：

```bash
make swagger
```

### 新增 API 端點

1. 在 handler 函數上方新增 Swagger 註解
2. 執行 `make swagger` 生成文檔
3. 重新編譯並啟動服務

### 測試

```bash
# 執行測試腳本
./scripts/test-span-names.sh

# 或使用 Makefile
make test-apis
```

## 相關文檔

- [Swagger API 完整文檔](SWAGGER_API.md)
- [架構改進說明](../ARCHITECTURE_IMPROVEMENT.md)
- [整合測試腳本](../test-integration.sh)

## 技術細節

### 依賴套件

- `github.com/swaggo/swag` - Swagger 文檔生成工具
- `github.com/swaggo/http-swagger` - Swagger UI HTTP handler

### 檔案結構

```
tempo-otlp-trace-demo/
├── docs/                    # Swagger 生成的文檔
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
├── handlers/
│   ├── spannames.go        # 新增的 span names handler
│   └── sourcecode.go       # 更新的 source code handler
├── scripts/
│   └── test-span-names.sh  # 測試腳本
├── Makefile                # 新增 swagger 相關指令
└── SWAGGER_API.md          # Swagger 完整文檔
```

## 效能考量

- Span names API 從記憶體中讀取，回應速度快
- Swagger UI 為靜態資源，不影響 API 效能
- 建議在生產環境中考慮快取策略

## 安全建議

在生產環境中：

1. 考慮為 Swagger UI 新增認證機制
2. 使用環境變數控制 Swagger UI 的啟用/停用
3. 設定適當的 CORS 策略
4. 限制 API 存取頻率

## 問題排查

### Swagger UI 顯示空白

確保已執行 `make swagger` 生成文檔。

### API 回傳 404

檢查服務是否正確啟動：

```bash
curl http://localhost:8080/health
```

### 映射資料不正確

重新載入映射：

```bash
curl -X POST http://localhost:8080/api/mappings/reload
```

## 下一步

- 探索 [Swagger UI](http://localhost:8080/swagger/) 的所有功能
- 查看 [完整 API 文檔](SWAGGER_API.md)
- 執行 [整合測試](../test-integration.sh)
