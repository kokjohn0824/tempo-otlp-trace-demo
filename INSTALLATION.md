# 安裝和設定指南

本指南將幫助你快速設定和開始使用 Tempo OTLP Trace Demo 專案。

## 📋 前置需求

### 必要工具

在開始之前，請確保你的系統已安裝以下工具：

#### 1. Go (1.21 或更高版本)

**檢查是否已安裝**:
```bash
go version
```

**安裝方式**:
- macOS: `brew install go`
- Linux: 參考 [官方文件](https://golang.org/doc/install)
- Windows: 下載安裝程式從 [golang.org](https://golang.org/dl/)

#### 2. Docker

**檢查是否已安裝**:
```bash
docker --version
```

**安裝方式**:
- macOS: 安裝 [Docker Desktop for Mac](https://www.docker.com/products/docker-desktop)
- Linux: 參考 [官方文件](https://docs.docker.com/engine/install/)
- Windows: 安裝 [Docker Desktop for Windows](https://www.docker.com/products/docker-desktop)

#### 3. Docker Compose

**檢查是否已安裝**:
```bash
docker-compose --version
# 或
docker compose version
```

**安裝方式**:
- 通常隨 Docker Desktop 一起安裝
- Linux: 參考 [官方文件](https://docs.docker.com/compose/install/)

#### 4. Make

**檢查是否已安裝**:
```bash
make --version
```

**安裝方式**:
- macOS: 通常預裝，或執行 `xcode-select --install`
- Linux: `sudo apt-get install build-essential` (Debian/Ubuntu)
- Windows: 安裝 [Make for Windows](http://gnuwin32.sourceforge.net/packages/make.htm)

#### 5. Git

**檢查是否已安裝**:
```bash
git --version
```

**安裝方式**:
- macOS: `brew install git` 或 `xcode-select --install`
- Linux: `sudo apt-get install git`
- Windows: 下載從 [git-scm.com](https://git-scm.com/)

### 選用工具

這些工具不是必需的，但建議安裝以獲得更好的開發體驗：

#### 1. curl 和 jq

用於 API 測試腳本。

**檢查是否已安裝**:
```bash
curl --version
jq --version
```

**安裝方式**:
- macOS: `brew install curl jq`
- Linux: `sudo apt-get install curl jq`
- Windows: 參考各工具的官方網站

#### 2. golangci-lint

用於程式碼檢查。

**安裝方式**:
```bash
# macOS
brew install golangci-lint

# Linux
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

# 或使用 Go
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

## 🚀 快速安裝

### 步驟 1: Clone 專案

```bash
git clone https://github.com/your-username/tempo-otlp-trace-demo.git
cd tempo-otlp-trace-demo
```

### 步驟 2: 檢查依賴

```bash
make check-deps
```

如果看到 "✓ 所有依賴工具已安裝"，表示你可以繼續。如果有錯誤，請根據上面的說明安裝缺少的工具。

### 步驟 3: 安裝 Go 依賴

```bash
make install-deps
```

### 步驟 4: 啟動服務

```bash
make up
```

這會啟動所有必要的服務：
- Go 應用程式 (port 8080)
- OpenTelemetry Collector (port 4317, 4318)
- Grafana Tempo (port 3200)
- Grafana (port 3000)

### 步驟 5: 驗證安裝

```bash
make health
```

你應該看到所有服務都顯示 "✓ OK"。

### 步驟 6: 執行測試

```bash
make test-apis
```

### 步驟 7: 開啟 Grafana

```bash
make open-grafana
```

或手動開啟瀏覽器訪問: http://localhost:3000

## 🔧 詳細安裝步驟

### 選項 1: 使用 Docker (推薦)

這是最簡單的方式，所有服務都在容器中運行。

```bash
# 1. Clone 專案
git clone https://github.com/your-username/tempo-otlp-trace-demo.git
cd tempo-otlp-trace-demo

# 2. 檢查依賴
make check-deps

# 3. 啟動所有服務
make up

# 4. 檢查健康狀態
make health

# 5. 執行測試
make test-apis

# 6. 查看日誌
make logs-app

# 7. 開啟 Grafana
make open-grafana
```

### 選項 2: 本地開發模式

在本地運行應用程式，其他服務在 Docker 中運行。

```bash
# 1. Clone 專案
git clone https://github.com/your-username/tempo-otlp-trace-demo.git
cd tempo-otlp-trace-demo

# 2. 安裝依賴
make install-deps

# 3. 啟動基礎設施（不含應用程式）
make infra-up

# 4. 在本地編譯並運行應用程式
make run

# 5. 在另一個終端執行測試
make test-quick
```

### 選項 3: 完整開發環境

適合需要頻繁修改程式碼的開發者。

```bash
# 1. Clone 專案
git clone https://github.com/your-username/tempo-otlp-trace-demo.git
cd tempo-otlp-trace-demo

# 2. 安裝依賴
make install-deps

# 3. 啟動開發模式
make dev

# 這會自動：
# - 啟動基礎設施
# - 編譯應用程式
# - 運行應用程式
# - 檢查健康狀態
```

## 🐛 常見問題排查

### 問題 1: Docker 啟動失敗

**錯誤訊息**: "Cannot connect to the Docker daemon"

**解決方法**:
1. 確保 Docker Desktop 正在運行
2. 檢查 Docker 狀態: `docker ps`
3. 重啟 Docker Desktop

### 問題 2: Port 已被佔用

**錯誤訊息**: "port is already allocated"

**解決方法**:
1. 檢查哪個程式佔用了 port:
   ```bash
   # macOS/Linux
   lsof -i :8080
   lsof -i :3000
   
   # Windows
   netstat -ano | findstr :8080
   ```

2. 停止佔用 port 的程式，或修改 `docker-compose.yml` 使用不同的 port

### 問題 3: Go 依賴下載失敗

**錯誤訊息**: "go: downloading ... timeout"

**解決方法**:
1. 檢查網路連線
2. 設定 Go proxy:
   ```bash
   export GOPROXY=https://proxy.golang.org,direct
   ```
3. 重試: `make install-deps`

### 問題 4: 權限問題

**錯誤訊息**: "permission denied"

**解決方法**:
1. 確保你有執行權限:
   ```bash
   chmod +x scripts/test-apis.sh
   ```

2. 如果是 Docker 權限問題:
   ```bash
   # Linux
   sudo usermod -aG docker $USER
   # 登出後重新登入
   ```

### 問題 5: 服務無法啟動

**症狀**: `make health` 顯示服務失敗

**解決方法**:
1. 查看日誌:
   ```bash
   make logs
   ```

2. 檢查服務狀態:
   ```bash
   make ps
   ```

3. 重啟服務:
   ```bash
   make restart
   ```

4. 如果還是失敗，完全重置:
   ```bash
   make down-volumes
   make up
   ```

## 📦 驗證安裝

執行以下指令確保一切正常：

```bash
# 1. 檢查所有服務都在運行
make ps

# 2. 檢查健康狀態
make health

# 3. 測試應用程式
curl http://localhost:8080/health

# 4. 執行 API 測試
make test-quick

# 5. 檢查 Grafana
curl http://localhost:3000/api/health
```

如果所有指令都成功執行，恭喜你！安裝完成。

## 🎓 下一步

現在你已經成功安裝了專案，可以：

1. **閱讀文件**:
   - [README.md](README.md) - 專案概覽
   - [QUICK_REFERENCE.md](QUICK_REFERENCE.md) - 快速參考
   - [MAKEFILE_GUIDE.md](MAKEFILE_GUIDE.md) - 詳細指南

2. **探索 API**:
   - 訪問 http://localhost:8080 查看 API 文件
   - 執行 `make test-apis` 查看所有 endpoints

3. **查看 Traces**:
   - 開啟 Grafana: http://localhost:3000
   - 前往 Explore → 選擇 Tempo
   - 搜尋 service name: `trace-demo-service`

4. **開始開發**:
   - 閱讀 [CONTRIBUTING.md](CONTRIBUTING.md)
   - 執行 `make dev` 啟動開發環境

## 🔄 更新專案

如果你已經安裝過專案，要更新到最新版本：

```bash
# 1. 拉取最新程式碼
git pull origin master

# 2. 更新依賴
make install-deps

# 3. 重建映像
make docker-build

# 4. 重啟服務
make restart

# 5. 驗證
make health
```

## 🗑️ 解除安裝

如果你想完全移除專案：

```bash
# 1. 停止所有服務
make down

# 2. 刪除 volumes
make down-volumes

# 3. 刪除 Docker 映像
docker rmi trace-demo-app:latest

# 4. 刪除專案目錄
cd ..
rm -rf tempo-otlp-trace-demo
```

## 💡 提示和技巧

### 提示 1: 使用別名

在你的 shell 配置文件（如 `.bashrc` 或 `.zshrc`）中添加別名：

```bash
alias tempo-up='cd ~/path/to/tempo-otlp-trace-demo && make up'
alias tempo-down='cd ~/path/to/tempo-otlp-trace-demo && make down'
alias tempo-logs='cd ~/path/to/tempo-otlp-trace-demo && make logs-app'
```

### 提示 2: 使用 tmux 或 screen

在多個終端視窗中同時查看不同的日誌：

```bash
# 終端 1
make logs-app

# 終端 2
make logs-collector

# 終端 3
make logs-tempo
```

### 提示 3: 自動啟動

在系統啟動時自動啟動服務（macOS 範例）：

```bash
# 創建 LaunchAgent
cat > ~/Library/LaunchAgents/com.tempo-demo.plist <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.tempo-demo</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/bin/make</string>
        <string>up</string>
    </array>
    <key>WorkingDirectory</key>
    <string>/path/to/tempo-otlp-trace-demo</string>
    <key>RunAtLoad</key>
    <true/>
</dict>
</plist>
EOF

# 載入
launchctl load ~/Library/LaunchAgents/com.tempo-demo.plist
```

## 📞 獲取幫助

如果你遇到問題：

1. 查看 [README.md](README.md) 和 [QUICK_REFERENCE.md](QUICK_REFERENCE.md)
2. 查看 [MAKEFILE_GUIDE.md](MAKEFILE_GUIDE.md) 的問題排查章節
3. 搜尋現有的 Issues
4. 建立新的 Issue 描述你的問題

## ✅ 安裝檢查清單

- [ ] Go 已安裝 (1.21+)
- [ ] Docker 已安裝並運行
- [ ] Docker Compose 已安裝
- [ ] Make 已安裝
- [ ] Git 已安裝
- [ ] 專案已 clone
- [ ] `make check-deps` 通過
- [ ] `make install-deps` 成功
- [ ] `make up` 成功
- [ ] `make health` 所有服務正常
- [ ] `make test-apis` 通過
- [ ] Grafana 可以訪問 (http://localhost:3000)
- [ ] 應用程式可以訪問 (http://localhost:8080)

---

**恭喜！你已經成功安裝 Tempo OTLP Trace Demo！** 🎉

開始探索: `make help`
