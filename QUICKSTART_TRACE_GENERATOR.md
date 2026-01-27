# Trace Generator - 快速入門

## 🚀 立即開始

### 1. 啟動所有服務

```bash
cd /Users/alexchang/dev/rag-slow/tempo-otlp-trace-demo
docker-compose -f docker-compose-deploy.yml up -d
```

### 2. 查看日誌

```bash
# 即時查看容器日誌
docker logs -f trace-generator

# 或查看日誌檔案
tail -f trace-generator/logs/trace-generator.log
```

### 3. 驗證運行

預期看到類似輸出：

```
2026/01/23 10:30:00 [INFO] Trace generator started
2026/01/23 10:30:00 [INFO] Target URL: http://trace-demo-app:8080
2026/01/23 10:30:00 [INFO] Interval: 30s
2026/01/23 10:30:00 [INFO] Starting API call cycle
2026/01/23 10:30:00 [INFO] API order succeeded (took 850ms)
2026/01/23 10:30:01 [INFO] API user succeeded (took 150ms)
2026/01/23 10:30:02 [INFO] API report succeeded (took 2.3s)
...
2026/01/23 10:30:07 [INFO] Cycle completed: 6 succeeded, 0 failed
```

## 📊 查看 Traces

1. 開啟 Grafana: http://localhost:3000
2. 前往 Explore
3. 選擇 Tempo 資料源
4. 搜尋 traces（應該會看到持續產生的新 traces）

## ⚙️ 常用操作

### 停止 Trace Generator

```bash
docker-compose -f docker-compose-deploy.yml stop trace-generator
```

### 重新啟動

```bash
docker-compose -f docker-compose-deploy.yml restart trace-generator
```

### 停止所有服務

```bash
docker-compose -f docker-compose-deploy.yml down
```

### 重新建構

```bash
docker-compose -f docker-compose-deploy.yml up -d --build trace-generator
```

## 🔧 調整配置

編輯 `docker-compose-deploy.yml` 中的環境變數：

```yaml
trace-generator:
  environment:
    - INTERVAL_SECONDS=60        # 改為每分鐘執行
    - ENABLED_APIS=order,user    # 只啟用特定 API
    - TIMEOUT_SECONDS=60         # 延長超時時間
```

然後重新啟動：

```bash
docker-compose -f docker-compose-deploy.yml up -d --force-recreate trace-generator
```

## 🔍 故障排除

### 問題：無法連接到 trace-demo-app

```bash
# 檢查 trace-demo-app 是否運行
docker ps | grep trace-demo-app

# 檢查網路連接
docker exec trace-generator ping -c 3 trace-demo-app
```

### 問題：API 呼叫失敗

```bash
# 查看詳細日誌
docker logs trace-generator | grep ERROR
```

### 問題：日誌檔案未建立

```bash
# 檢查 volume 掛載
docker inspect trace-generator | grep Mounts -A 10

# 確認目錄存在
ls -la trace-generator/logs/
```

## 📚 更多資訊

- 詳細說明：[README.md](trace-generator/README.md)
- 實作文件：[IMPLEMENTATION.md](trace-generator/IMPLEMENTATION.md)
- 完整總結：[TRACE_GENERATOR_SUMMARY.md](TRACE_GENERATOR_SUMMARY.md)

## ✨ 就是這麼簡單！

Trace Generator 現在會每 30 秒自動呼叫所有 API 端點，持續產生 traces 供 Tempo 收集和分析。
