# 貢獻指南

感謝你對本專案的興趣！本文件說明如何為專案做出貢獻。

## 目錄

- [開發環境設定](#開發環境設定)
- [開發流程](#開發流程)
- [程式碼規範](#程式碼規範)
- [提交規範](#提交規範)
- [測試要求](#測試要求)
- [Pull Request 流程](#pull-request-流程)

## 開發環境設定

### 前置需求

確保你已安裝以下工具：

- Go 1.21 或更高版本
- Docker 和 Docker Compose
- Make
- Git
- curl 和 jq (用於測試)

### 檢查依賴

```bash
make check-deps
```

### 初始設定

1. Fork 並 clone 專案：

```bash
git clone https://github.com/your-username/tempo-otlp-trace-demo.git
cd tempo-otlp-trace-demo
```

2. 安裝依賴：

```bash
make install-deps
```

3. 啟動開發環境：

```bash
make dev
```

4. 驗證環境：

```bash
make health
make test-quick
```

## 開發流程

### 1. 創建功能分支

```bash
git checkout -b feature/your-feature-name
```

### 2. 進行開發

使用開發模式啟動服務：

```bash
make dev
```

在另一個終端視窗查看日誌：

```bash
make logs-app
```

### 3. 測試你的更改

#### 格式化程式碼

```bash
make fmt
```

#### 執行靜態檢查

```bash
make vet
```

#### 執行單元測試

```bash
make test
```

#### 執行 API 測試

```bash
make test-apis
```

### 4. 提交前檢查

在提交前，確保所有檢查都通過：

```bash
make ci
```

這會執行：
- 程式碼格式化
- 靜態檢查
- 單元測試
- Docker 映像建立

## 程式碼規範

### Go 程式碼規範

1. **格式化**：使用 `gofmt` 格式化程式碼（執行 `make fmt`）

2. **命名規範**：
   - 使用駝峰式命名（camelCase）
   - 導出的函數和變數使用大寫開頭（PascalCase）
   - 私有函數和變數使用小寫開頭（camelCase）

3. **註解**：
   - 所有導出的函數都應該有註解
   - 註解應該以函數名稱開頭
   - 複雜的邏輯應該有解釋性註解

4. **錯誤處理**：
   - 不要忽略錯誤
   - 使用有意義的錯誤訊息
   - 適當地包裝錯誤（使用 `fmt.Errorf`）

### 範例

```go
// CreateOrder 建立一個新訂單並返回訂單 ID
// 如果驗證失敗或資料庫操作失敗，返回錯誤
func CreateOrder(ctx context.Context, req OrderRequest) (string, error) {
    // 驗證請求
    if err := validateOrderRequest(req); err != nil {
        return "", fmt.Errorf("invalid order request: %w", err)
    }
    
    // 建立訂單
    orderID, err := createOrderInDB(ctx, req)
    if err != nil {
        return "", fmt.Errorf("failed to create order: %w", err)
    }
    
    return orderID, nil
}
```

## 提交規範

### 提交訊息格式

使用以下格式撰寫提交訊息：

```
<type>(<scope>): <subject>

<body>

<footer>
```

#### Type

- `feat`: 新功能
- `fix`: 錯誤修復
- `docs`: 文件更新
- `style`: 程式碼格式化（不影響功能）
- `refactor`: 重構（不是新功能也不是錯誤修復）
- `test`: 測試相關
- `chore`: 建構流程或輔助工具的變動

#### 範例

```
feat(handlers): add new batch processing endpoint

Add a new endpoint /api/batch/process that allows processing
multiple items in a single request. Each item is processed
with its own span for better tracing.

Closes #123
```

## 測試要求

### 單元測試

- 所有新功能都應該有單元測試
- 測試覆蓋率應該保持在 80% 以上
- 使用表格驅動測試（table-driven tests）

範例：

```go
func TestCreateOrder(t *testing.T) {
    tests := []struct {
        name    string
        req     OrderRequest
        wantErr bool
    }{
        {
            name: "valid order",
            req: OrderRequest{
                UserID:    "user123",
                ProductID: "prod456",
                Quantity:  2,
                Price:     99.99,
            },
            wantErr: false,
        },
        {
            name: "invalid quantity",
            req: OrderRequest{
                UserID:    "user123",
                ProductID: "prod456",
                Quantity:  0,
                Price:     99.99,
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := CreateOrder(context.Background(), tt.req)
            if (err != nil) != tt.wantErr {
                t.Errorf("CreateOrder() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### 整合測試

新增或修改 API endpoint 時：

1. 更新 `scripts/test-apis.sh`
2. 確保測試腳本能正常執行：

```bash
make test-apis
```

### 執行所有測試

```bash
# 單元測試
make test

# 測試覆蓋率
make test-coverage

# API 測試
make test-apis

# 完整 CI 流程
make ci
```

## Pull Request 流程

### 1. 確保你的分支是最新的

```bash
git checkout master
git pull origin master
git checkout feature/your-feature-name
git rebase master
```

### 2. 執行完整檢查

```bash
make ci
```

### 3. 推送你的分支

```bash
git push origin feature/your-feature-name
```

### 4. 建立 Pull Request

在 GitHub 上建立 Pull Request，並確保：

- 標題清楚描述變更
- 描述中包含：
  - 變更的原因
  - 變更的內容
  - 測試方法
  - 相關的 issue 編號（如果有）
- 所有 CI 檢查都通過

### Pull Request 模板

```markdown
## 描述

簡要描述這個 PR 的目的和變更內容。

## 變更類型

- [ ] 新功能
- [ ] 錯誤修復
- [ ] 重構
- [ ] 文件更新
- [ ] 其他（請說明）

## 測試

描述你如何測試這些變更：

- [ ] 單元測試已通過
- [ ] API 測試已通過
- [ ] 手動測試已完成

測試步驟：
1. ...
2. ...

## 檢查清單

- [ ] 程式碼已格式化（`make fmt`）
- [ ] 靜態檢查已通過（`make vet`）
- [ ] 所有測試都通過（`make test`）
- [ ] 已更新相關文件
- [ ] 提交訊息符合規範
- [ ] CI 檢查已通過

## 相關 Issue

Closes #(issue number)

## 截圖（如果適用）

## 其他說明
```

## 開發技巧

### 快速迭代

使用開發模式可以快速測試變更：

```bash
# 終端 1: 啟動基礎設施
make infra-up

# 終端 2: 運行應用程式（修改後重新執行）
make run

# 終端 3: 查看日誌
make logs-app

# 終端 4: 執行測試
make test-quick
```

### 除錯技巧

1. **查看詳細日誌**：

```bash
make logs-app
```

2. **檢查服務健康狀態**：

```bash
make health
```

3. **在 Grafana 中查看 traces**：

```bash
make open-grafana
```

4. **使用自訂參數測試**：

```bash
curl "http://localhost:8080/api/simulate?depth=5&breadth=2&duration=100&variance=0.5"
```

### 效能分析

```bash
# 執行效能測試
make bench

# 生成 CPU profile
go test -cpuprofile=cpu.prof -bench=.

# 分析 profile
go tool pprof cpu.prof
```

## 常見問題

### Q: 如何重置開發環境？

```bash
make down-volumes
make up
```

### Q: 測試失敗怎麼辦？

1. 查看錯誤訊息
2. 檢查日誌：`make logs`
3. 確保服務都在運行：`make ps`
4. 重啟服務：`make restart`

### Q: 如何添加新的 API endpoint？

1. 在 `handlers/` 目錄下建立新的處理器
2. 在 `main.go` 中註冊路由
3. 更新 `scripts/test-apis.sh` 添加測試
4. 更新 README.md 文件
5. 執行測試確保一切正常

### Q: 如何更新依賴？

```bash
go get -u ./...
make tidy
make test
```

## 獲取幫助

如果你遇到問題：

1. 查看 [README.md](README.md) 和 [INSTALLATION.md](INSTALLATION.md)
2. 查看 [MAKEFILE_GUIDE.md](MAKEFILE_GUIDE.md) 和 [QUICK_REFERENCE.md](QUICK_REFERENCE.md)
3. 搜尋現有的 Issues
4. 建立新的 Issue 描述你的問題

## 行為準則

- 尊重所有貢獻者
- 建設性地提供反饋
- 專注於問題本身，而不是個人
- 歡迎新手貢獻者

## 授權

提交 Pull Request 即表示你同意你的貢獻將以專案的授權條款（MIT License）發布。

---

再次感謝你的貢獻！🎉
