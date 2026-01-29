# Changelog

本文件記錄專案的所有重要變更。

格式基於 [Keep a Changelog](https://keepachangelog.com/zh-TW/1.0.0/)，
版本號遵循 [Semantic Versioning](https://semver.org/lang/zh-TW/)。

## [Unreleased]

### Added

- **原始碼分析 API**
  - `GET /api/source-code` - 根據 span ID 和 trace ID 獲取原始碼
  - `GET /api/mappings` - 查詢所有原始碼映射
  - `POST /api/mappings` - 新增/更新原始碼映射
  - `DELETE /api/mappings` - 刪除原始碼映射
  - `POST /api/mappings/reload` - 重新載入映射表

- **Tempo 查詢功能** (`tracing/tempo.go`)
  - 支援透過 trace ID 查詢完整的 trace 資訊
  - 自動解析 span 資料和關聯關係

- **Makefile 建構系統**
  - 開發工作流程指令 (dev, run, build)
  - Docker 操作指令 (up, down, restart)
  - 測試指令 (test, test-coverage, test-apis)
  - 日誌和監控指令 (logs, health, ps)

- **CI/CD**
  - `.github/workflows/ci.yml` - GitHub Actions CI 流程
  - `.github/workflows/deploy.yml` - GitHub Actions 部署流程

- **Swagger API 文檔**
  - 互動式 API 文檔介面
  - 訪問 http://localhost:8080/swagger/

### Changed

- 更新 README.md 添加 Makefile 使用說明
- 執行 `go fmt` 格式化所有 Go 程式碼

## 使用指南

```bash
make help        # 查看所有可用指令
make up          # 啟動所有服務
make health      # 檢查健康狀態
make test-apis   # 執行 API 測試
make dev         # 啟動開發環境
```
