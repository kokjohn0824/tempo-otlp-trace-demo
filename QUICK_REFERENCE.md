# Makefile 快速參考

快速查找常用的 Makefile 指令。

## 🚀 快速開始

```bash
make up          # 啟動所有服務
make health      # 檢查健康狀態
make test-apis   # 執行 API 測試
make open-grafana # 開啟 Grafana
```

## 📦 開發

| 指令 | 說明 |
|------|------|
| `make dev` | 開發模式（啟動基礎設施 + 本地運行應用） |
| `make run` | 在本地運行應用程式 |
| `make build` | 編譯應用程式 |
| `make build-local` | 編譯本地版本 |
| `make fmt` | 格式化程式碼 |
| `make vet` | 靜態檢查 |
| `make lint` | Lint 檢查 |

## 🐳 Docker

| 指令 | 說明 |
|------|------|
| `make up` | 啟動所有服務 |
| `make down` | 停止所有服務 |
| `make restart` | 重啟服務 |
| `make docker-build` | 建立 Docker 映像 |
| `make docker-push` | 推送映像 |
| `make ps` | 查看服務狀態 |

## 📝 日誌

| 指令 | 說明 |
|------|------|
| `make logs` | 所有服務日誌 |
| `make logs-app` | 應用程式日誌 |
| `make logs-collector` | OTel Collector 日誌 |
| `make logs-tempo` | Tempo 日誌 |
| `make logs-grafana` | Grafana 日誌 |

## 🧪 測試

| 指令 | 說明 |
|------|------|
| `make test` | 單元測試 |
| `make test-coverage` | 測試 + 覆蓋率報告 |
| `make test-apis` | API 測試 |
| `make test-quick` | 快速 API 測試 |
| `make bench` | 效能測試 |

## 🔍 監控

| 指令 | 說明 |
|------|------|
| `make health` | 健康檢查 |
| `make ps` | 服務狀態 |
| `make open-grafana` | 開啟 Grafana |
| `make open-app` | 開啟應用程式 |

## 🧹 清理

| 指令 | 說明 |
|------|------|
| `make clean` | 清理編譯產物 |
| `make clean-all` | 完全清理（含 Docker） |
| `make down-volumes` | 停止並刪除 volumes |

## 🚢 部署

| 指令 | 說明 |
|------|------|
| `make ci` | CI 流程 |
| `make deploy` | 建立並推送映像 |
| `make all` | 完整建立流程 |

## 💡 常用組合

### 第一次使用
```bash
make check-deps && make install-deps && make up
```

### 開始開發
```bash
make dev
```

### 提交前檢查
```bash
make fmt && make vet && make test
```

### 完整測試
```bash
make ci && make test-apis
```

### 重置環境
```bash
make down-volumes && make up
```

### 查看問題
```bash
make health && make logs-app
```

## 🔧 環境變數

```bash
# 自訂 Docker Registry
make deploy DOCKER_REGISTRY=myregistry.com DOCKER_TAG=v1.0.0

# 自訂 Port
make run PORT=9090

# 測試遠端服務
make test-apis BASE_URL=http://remote:8080
```

## 📚 更多資訊

- **完整指令列表**：`make help`
- **詳細使用指南**：[MAKEFILE_GUIDE.md](MAKEFILE_GUIDE.md)
- **安裝說明**：[INSTALLATION.md](INSTALLATION.md)
- **貢獻指南**：[CONTRIBUTING.md](CONTRIBUTING.md)
- **專案說明**：[README.md](README.md)
