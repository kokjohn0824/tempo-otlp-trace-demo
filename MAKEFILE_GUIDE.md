# Makefile 使用指南

本文件提供 Makefile 的詳細使用說明和常見工作流程範例。

## 目錄

- [快速開始](#快速開始)
- [開發工作流程](#開發工作流程)
- [測試工作流程](#測試工作流程)
- [部署工作流程](#部署工作流程)
- [常見問題](#常見問題)

## 快速開始

### 第一次使用

```bash
# 1. 檢查依賴工具是否已安裝
make check-deps

# 2. 安裝 Go 依賴
make install-deps

# 3. 啟動所有服務
make up

# 4. 檢查服務健康狀態
make health

# 5. 執行 API 測試
make test-apis

# 6. 在瀏覽器中開啟 Grafana
make open-grafana
```

### 日常開發

```bash
# 啟動開發環境
make dev

# 在另一個終端視窗查看日誌
make logs-app

# 執行測試
make test-quick
```

## 開發工作流程

### 工作流程 1: 完整的本地開發

```bash
# 1. 格式化程式碼
make fmt

# 2. 執行靜態檢查
make vet

# 3. 執行單元測試
make test

# 4. 啟動開發環境 (基礎設施 + 應用程式)
make dev
```

### 工作流程 2: Docker 開發

```bash
# 1. 建立 Docker 映像
make docker-build

# 2. 啟動所有服務 (包含應用程式)
make up

# 3. 查看應用程式日誌
make logs-app

# 4. 執行 API 測試
make test-apis

# 5. 停止服務
make down
```

### 工作流程 3: 只修改程式碼

```bash
# 1. 啟動基礎設施 (不含應用程式)
make infra-up

# 2. 在本地運行應用程式 (會自動編譯)
make run

# 修改程式碼後，按 Ctrl+C 停止，然後重新執行
make run
```

## 測試工作流程

### 單元測試

```bash
# 執行所有單元測試
make test

# 執行測試並生成覆蓋率報告
make test-coverage

# 執行效能測試
make bench
```

### API 測試

```bash
# 完整 API 測試 (包含所有 endpoints)
make test-apis

# 快速測試 (減少等待時間)
make test-quick

# 測試特定環境
make test-apis BASE_URL=http://staging-server:8080
```

### 整合測試

```bash
# 1. 啟動所有服務
make up

# 2. 等待服務就緒
make health

# 3. 執行 API 測試
make test-apis

# 4. 查看 Grafana 中的 traces
make open-grafana

# 5. 清理
make down
```

## 部署工作流程

### 工作流程 1: 部署到 Docker Registry

```bash
# 1. 執行 CI 檢查
make ci

# 2. 建立並推送映像
make deploy DOCKER_REGISTRY=myregistry.com DOCKER_TAG=v1.0.0

# 3. 在目標環境執行
# ssh to-production-server
# docker-compose pull
# docker-compose up -d
```

### 工作流程 2: 本地完整建立

```bash
# 執行完整建立流程 (清理、安裝依賴、測試、建立)
make all
```

### 工作流程 3: CI/CD 流程

```bash
# 在 CI 環境中執行
make ci

# 這會執行:
# - make fmt (格式化)
# - make vet (靜態檢查)
# - make test (單元測試)
# - make docker-build (建立映像)
```

## 日誌和監控

### 查看日誌

```bash
# 查看所有服務日誌
make logs

# 查看特定服務日誌
make logs-app        # 應用程式
make logs-collector  # OTel Collector
make logs-tempo      # Tempo
make logs-grafana    # Grafana

# 查看服務狀態
make ps
```

### 健康檢查

```bash
# 檢查所有服務健康狀態
make health

# 輸出範例:
# 應用程式 (port 8080): ✓ OK
# OTel Collector (port 13133): ✓ OK
# Tempo (port 3200): ✓ OK
# Grafana (port 3000): ✓ OK
```

## 清理和維護

### 清理編譯產物

```bash
# 清理 Go 編譯產物
make clean

# 完全清理 (包含 Docker)
make clean-all
```

### 重啟服務

```bash
# 重啟所有服務
make restart

# 等同於
make down && make up
```

### 清理資料

```bash
# 停止服務並刪除所有資料 (會提示確認)
make down-volumes
```

### 整理依賴

```bash
# 整理 Go 依賴
make tidy
```

## 環境變數配置

### Docker Registry

```bash
# 設定自訂 Registry
export DOCKER_REGISTRY=myregistry.com
export DOCKER_TAG=v1.0.0

make docker-build
make docker-push
```

### 應用程式 Port

```bash
# 使用自訂 port
make run PORT=9090
```

### API 測試 URL

```bash
# 測試遠端服務
make test-apis BASE_URL=http://production:8080
```

## 常見使用情境

### 情境 1: 早上開始工作

```bash
# 1. 拉取最新程式碼
git pull

# 2. 安裝/更新依賴
make install-deps

# 3. 啟動開發環境
make dev

# 4. 在瀏覽器開啟 Grafana
make open-grafana
```

### 情境 2: 提交程式碼前

```bash
# 1. 格式化程式碼
make fmt

# 2. 執行檢查
make vet

# 3. 執行測試
make test

# 4. 如果都通過，提交程式碼
git add .
git commit -m "your message"
git push
```

### 情境 3: 除錯問題

```bash
# 1. 啟動服務
make up

# 2. 查看應用程式日誌
make logs-app

# 3. 在另一個終端執行測試
make test-apis

# 4. 檢查服務健康狀態
make health

# 5. 查看特定服務日誌
make logs-collector  # 如果是 tracing 問題
make logs-tempo      # 如果是 Tempo 問題
```

### 情境 4: 效能測試

```bash
# 1. 啟動服務
make up

# 2. 執行效能測試
make bench

# 3. 產生大量 traces
for i in {1..100}; do
  make test-quick &
done
wait

# 4. 在 Grafana 中分析結果
make open-grafana
```

### 情境 5: 準備發布

```bash
# 1. 執行完整 CI 流程
make ci

# 2. 標記版本
git tag v1.0.0
git push --tags

# 3. 建立並推送映像
make deploy DOCKER_REGISTRY=myregistry.com DOCKER_TAG=v1.0.0

# 4. 清理本地環境
make clean-all
```

## 常見問題

### Q: 服務啟動失敗怎麼辦？

```bash
# 1. 檢查 Docker 是否運行
docker ps

# 2. 查看服務狀態
make ps

# 3. 查看日誌找出問題
make logs

# 4. 嘗試完全重啟
make down
make up
```

### Q: 如何清理所有資料重新開始？

```bash
# 停止服務並刪除所有資料
make down-volumes

# 重新啟動
make up
```

### Q: 如何只重新建立應用程式映像？

```bash
# 重新建立映像
make docker-build

# 重啟應用程式容器
docker-compose up -d --force-recreate trace-demo-app
```

### Q: 如何在不同的 port 運行？

```bash
# 修改 docker-compose.yml 中的 port 映射
# 或使用環境變數
PORT=9090 make run
```

### Q: 測試腳本執行太慢怎麼辦？

```bash
# 使用快速測試模式
make test-quick

# 或手動設定等待時間
SLEEP_BETWEEN_CALLS=0.1 ./scripts/test-apis.sh
```

### Q: 如何查看測試覆蓋率？

```bash
# 生成覆蓋率報告
make test-coverage

# 會自動開啟 coverage.html
```

### Q: 如何在 CI/CD 中使用？

```yaml
# GitHub Actions 範例
- name: Run CI
  run: make ci

- name: Build and Push
  run: |
    make deploy DOCKER_REGISTRY=${{ secrets.REGISTRY }} \
                DOCKER_TAG=${{ github.sha }}
```

## 進階技巧

### 並行執行測試

```bash
# 在背景執行多個測試
make test-quick &
make test-quick &
make test-quick &
wait
```

### 自訂 Docker 映像名稱

```bash
# 修改 Makefile 中的變數
APP_NAME=my-custom-name make docker-build
```

### 監控資源使用

```bash
# 查看容器資源使用
docker stats

# 查看服務狀態
make ps
```

### 除錯 Docker 建立

```bash
# 查看建立過程
docker-compose build --no-cache --progress=plain
```

## 總結

Makefile 提供了一個統一的介面來管理專案的整個生命週期，從開發到測試到部署。記住：

- 使用 `make help` 查看所有可用指令
- 使用 `make dev` 快速開始開發
- 使用 `make ci` 在提交前檢查程式碼
- 使用 `make deploy` 部署到生產環境

## 📚 相關文件

- **[QUICK_REFERENCE.md](QUICK_REFERENCE.md)** - 快速參考卡片
- **[INSTALLATION.md](INSTALLATION.md)** - 安裝指南
- **[README.md](README.md)** - 專案說明
- **[CONTRIBUTING.md](CONTRIBUTING.md)** - 貢獻指南
