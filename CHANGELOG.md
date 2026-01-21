# Changelog

本文件記錄專案的所有重要變更。

格式基於 [Keep a Changelog](https://keepachangelog.com/zh-TW/1.0.0/)，
版本號遵循 [Semantic Versioning](https://semver.org/lang/zh-TW/)。

## [Unreleased]

### Added - 新增功能

#### 原始碼分析 API (Source Code Analysis API)
- 🆕 新增原始碼分析功能，可根據 span 資訊獲取對應的原始碼
  - `GET /api/source-code` - 根據 span ID 和 trace ID 獲取原始碼
  - `GET /api/mappings` - 查詢所有原始碼映射
  - `POST /api/mappings` - 新增/更新原始碼映射
  - `DELETE /api/mappings` - 刪除原始碼映射
  - `POST /api/mappings/reload` - 重新載入映射表

- 新增 Tempo 查詢功能 (`tracing/tempo.go`)
  - 支援透過 trace ID 查詢完整的 trace 資訊
  - 自動解析 span 資料和關聯關係
  - 提供 child spans 查詢功能

- 新增原始碼映射管理 (`handlers/sourcecode.go`)
  - 自動載入和儲存映射表
  - 支援動態更新映射
  - 讀取指定行數的原始碼

- 新增資料模型 (`models/request.go`)
  - `SourceCodeMapping` - 原始碼映射結構
  - `SourceCodeResponse` - API 回應結構
  - `ChildSpanInfo` - 子 span 資訊結構
  - `MappingRequest/Response` - 映射管理請求/回應

- 新增映射表設定檔 (`source_code_mappings.json`)
  - 預設包含所有現有 API endpoints 的映射
  - 支援自訂映射和動態更新

#### 文件系統
- 新增 `SOURCE_CODE_API.md` - 原始碼分析 API 完整文件
- 新增 `USAGE_EXAMPLE.md` - 原始碼分析 API 使用範例

#### 測試工具
- 新增 `scripts/test-source-code-api.sh` - 原始碼分析 API 測試腳本
  - 自動測試所有新增的 API endpoints
  - 驗證映射表功能
  - 測試錯誤處理
  - 提供手動測試指引

#### 功能特點
- ✅ 自動從 Tempo 查詢 span 資訊
- ✅ 根據 span name 自動映射到原始碼位置
- ✅ 回傳完整的原始碼和 metadata
- ✅ 包含 child spans 資訊和 duration 分析
- ✅ 支援動態更新映射表
- ✅ LLM 友善的 JSON 輸出格式
- ✅ 完整的錯誤處理和驗證

#### Makefile 和建構系統
- 新增完整的 Makefile，包含 40+ 個指令
  - 開發工作流程指令 (dev, run, build, build-local)
  - Docker 操作指令 (up, down, restart, docker-build, docker-push)
  - 測試指令 (test, test-coverage, test-apis, test-quick, bench)
  - 日誌和監控指令 (logs, logs-*, health, ps)
  - 清理和維護指令 (clean, clean-all, tidy)
  - CI/CD 指令 (ci, deploy, all)
  - 工具指令 (check-deps, install-deps, fmt, vet, lint)
  - 便利指令 (open-grafana, open-app)

#### 文件系統
- 新增 `INSTALLATION.md` - 詳細的安裝和設定指南
  - 前置需求檢查清單
  - 三種安裝選項（Docker、本地開發、完整開發）
  - 常見問題排查
  - 驗證安裝步驟
  - 提示和技巧

- 新增 `MAKEFILE_GUIDE.md` - Makefile 詳細使用指南
  - 快速開始教學
  - 各種工作流程範例（開發、測試、部署）
  - 常見使用情境
  - 問題排查指南
  - 進階技巧

- 新增 `QUICK_REFERENCE.md` - Makefile 快速參考卡片
  - 指令分類速查表
  - 常用組合指令
  - 環境變數說明

- 新增 `CONTRIBUTING.md` - 貢獻指南
  - 開發環境設定
  - 開發流程
  - 程式碼規範
  - 提交規範
  - Pull Request 流程
  - 測試要求

- 新增 `CHANGELOG.md` - 變更日誌（本文件）

#### CI/CD
- 新增 `.github/workflows/ci.yml` - GitHub Actions CI 流程
  - 自動執行測試
  - 建立 Docker 映像
  - 整合測試
  - 上傳測試覆蓋率報告

- 新增 `.github/workflows/deploy.yml` - GitHub Actions 部署流程
  - 標籤觸發的自動部署
  - 建立並推送 Docker 映像
  - 自動建立 GitHub Release

### Changed - 變更

#### 文件更新
- 更新 `README.md`
  - 添加文件導覽章節
  - 添加 Makefile 使用說明
  - 重構快速開始章節，提供 Makefile 和手動兩種方式
  - 添加 Makefile 指令參考表
  - 更新本地開發模式說明
  - 更新停止服務說明

#### 配置文件
- 更新 `.gitignore`
  - 添加建立產物忽略規則 (`bin/`, `coverage.out`, `coverage.html`)
  - 添加測試產物忽略規則 (`*.test`, `*.prof`)

#### 依賴管理
- 執行 `go mod tidy` 整理依賴
- 更新 `go.mod` 和 `go.sum`

#### 程式碼格式化
- 執行 `go fmt` 格式化所有 Go 程式碼
- 格式化 `models/request.go`

### Improved - 改進

#### 開發體驗
- 統一的指令介面，減少記憶負擔
- 彩色輸出提供清晰的視覺反饋
- 自動依賴檢查，減少環境問題
- 整合的健康檢查，快速診斷問題
- 簡化的測試流程，提高測試效率

#### 文件品質
- 完整的文件體系，從快速參考到詳細指南
- 清晰的範例和使用情境
- 詳細的問題排查說明
- 中文化的文件，更易於理解

#### 自動化
- 自動化的建構流程
- 自動化的測試流程
- 自動化的部署流程
- CI/CD 整合

#### 程式碼品質
- 統一的程式碼格式化
- 自動的靜態檢查
- 測試覆蓋率報告
- Lint 檢查支援

### Technical Details - 技術細節

#### Makefile 特點
- 使用 ANSI 顏色碼提供視覺反饋
- 支援環境變數配置
- 自動依賴檢查
- 錯誤處理和確認提示
- 跨平台支援（macOS, Linux）
- 並行執行支援
- 清晰的幫助訊息

#### 目錄結構變更
```
新增:
├── .github/
│   └── workflows/
│       ├── ci.yml
│       └── deploy.yml
├── CHANGELOG.md
├── CONTRIBUTING.md
├── INSTALLATION.md
├── Makefile
├── MAKEFILE_GUIDE.md
├── QUICK_REFERENCE.md
├── SOURCE_CODE_API.md
└── USAGE_EXAMPLE.md

修改:
├── .gitignore
├── README.md
├── go.mod
├── go.sum
└── models/request.go
```

## 使用指南

### 查看所有可用指令
```bash
make help
```

### 快速開始
```bash
make up          # 啟動所有服務
make health      # 檢查健康狀態
make test-apis   # 執行 API 測試
```

### 開發模式
```bash
make dev         # 啟動開發環境
```

### CI 檢查
```bash
make ci          # 執行 CI 流程
```

### 部署
```bash
make deploy DOCKER_REGISTRY=myregistry.com DOCKER_TAG=v1.0.0
```

## 遷移指南

如果你之前使用手動指令，現在可以使用對應的 Makefile 指令：

| 舊指令 | 新指令 |
|--------|--------|
| `docker-compose up -d` | `make up` |
| `docker-compose down` | `make down` |
| `docker-compose logs -f trace-demo-app` | `make logs-app` |
| `./scripts/test-apis.sh` | `make test-apis` |
| `go build -o app .` | `make build-local` |
| `go fmt ./...` | `make fmt` |
| `go test ./...` | `make test` |
| `docker build -t trace-demo-app .` | `make docker-build` |

## 破壞性變更

無破壞性變更。所有原有的手動操作方式仍然可用。

## 已知問題

無已知問題。

## 未來計劃

### 短期
- [ ] 添加更多單元測試
- [ ] 添加效能測試基準
- [ ] 改進錯誤處理

### 中期
- [ ] 添加資料庫遷移支援
- [ ] 添加安全掃描
- [ ] 添加 API 文件生成

### 長期
- [ ] 添加監控儀表板
- [ ] 添加告警配置
- [ ] 支援多環境部署

## 貢獻者

感謝所有為本專案做出貢獻的人！

## 參考資源

- [Keep a Changelog](https://keepachangelog.com/zh-TW/1.0.0/)
- [Semantic Versioning](https://semver.org/lang/zh-TW/)
- [Make 文件](https://www.gnu.org/software/make/manual/)
- [GitHub Actions 文件](https://docs.github.com/en/actions)

---

**注意**: 本 CHANGELOG 遵循 [Keep a Changelog](https://keepachangelog.com/zh-TW/1.0.0/) 格式。
