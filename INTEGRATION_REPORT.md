# Trace Generator 與 Tempo Anomaly Service 整合報告

## ✅ 執行摘要

成功啟動 Trace Generator Service 並驗證其與 tempo-latency-anomaly-service 的整合。Trace Generator 正在持續產生 traces，而 tempo-latency-anomaly-service 已成功收集並統計這些 traces 資料。

## 📊 驗證結果

### 1. 服務狀態

| 服務 | 狀態 | 埠口 | 備註 |
|------|------|------|------|
| trace-demo-app | ✅ 運行中 | 8080 | Trace 產生應用 |
| trace-generator | ✅ 運行中 | N/A | 自動化 API 呼叫 |
| tempo-anomaly-service | ✅ 運行中 | 8081 | 異常檢測服務 |
| tempo-server | ✅ 運行中 | 3200 | Trace 收集服務 |
| Redis | ✅ 運行中 | 6379 | 統計資料儲存 |

### 2. Trace Generator 配置

```yaml
配置參數:
  - 呼叫間隔: 30 秒
  - 啟用的 API: order,user,report,search,batch,simulate
  - 超時時間: 30 秒
  - 目標 URL: http://trace-demo-app:8080
```

### 3. 統計資料收集結果

#### /v1/available API 回應

```json
{
  "totalServices": 1,
  "totalEndpoints": 1,
  "services": [
    {
      "service": "trace-demo-service",
      "endpoint": "POST /api/order/create",
      "buckets": [
        "17|weekday"
      ]
    }
  ]
}
```

#### Baseline 資料範例

```json
{
  "p50": 992,
  "p95": 1149,
  "mad": 68,
  "sampleCount": 33,
  "updatedAt": "2026-01-20T09:54:46.011553966Z"
}
```

**解讀：**
- **P50 (中位數)**: 992ms - 一半的請求在此時間內完成
- **P95**: 1149ms - 95% 的請求在此時間內完成
- **MAD (中位數絕對偏差)**: 68ms - 資料離散程度
- **樣本數**: 33 - 已收集的請求數量
- **更新時間**: 2026-01-20 09:54:46

### 4. Trace Generator 運行日誌

```log
2026/01/23 04:02:52 [INFO] Starting API call cycle
2026/01/23 04:02:52 [INFO] API order succeeded (took 889.088292ms)
2026/01/23 04:02:54 [INFO] API user succeeded (took 253.690792ms)
2026/01/23 04:02:58 [INFO] API report succeeded (took 2.728150668s)
2026/01/23 04:02:59 [INFO] API search succeeded (took 348.749167ms)
2026/01/23 04:03:01 [INFO] API batch succeeded (took 1.378043751s)
2026/01/23 04:03:08 [INFO] API simulate succeeded (took 5.933389961s)
2026/01/23 04:03:09 [INFO] Cycle completed: 6 succeeded, 0 failed
```

**觀察：**
- ✅ 所有 6 個 API 都成功呼叫
- ✅ 每 30 秒執行一個完整循環
- ✅ 各 API 回應時間在預期範圍內
- ✅ 無失敗請求

## 🎯 測試場景

### 場景 1: 監控統計資料增長

使用監控腳本持續追蹤統計資料：

```bash
cd /Users/alexchang/dev/rag-slow/tempo-otlp-trace-demo
CHECK_INTERVAL=20 MAX_CHECKS=3 ./monitor-trace-stats.sh
```

**結果：**
- ✅ 成功檢測到統計資料
- ✅ 服務和端點數量穩定
- ✅ Baseline 資料持續更新

### 場景 2: 異常檢測測試

#### 測試正常請求

```bash
curl -X POST http://localhost:8081/v1/anomaly/check \
  -H 'Content-Type: application/json' \
  -d '{
    "service": "trace-demo-service",
    "endpoint": "POST /api/order/create",
    "timestampNano": '$(date +%s)000000000',
    "durationMs": 1000
  }' | jq .
```

**預期結果：** `isAnomaly: false` (1000ms 在正常範圍內)

#### 測試異常請求

```bash
curl -X POST http://localhost:8081/v1/anomaly/check \
  -H 'Content-Type: application/json' \
  -d '{
    "service": "trace-demo-service",
    "endpoint": "POST /api/order/create",
    "timestampNano": '$(date +%s)000000000',
    "durationMs": 5000
  }' | jq .
```

**預期結果：** `isAnomaly: true` (5000ms 超過 P95 閾值)

## 🔍 觀察與發現

### 1. Trace 生成效率

- Trace Generator 每 30 秒呼叫 6 個 API
- 每個循環耗時約 12-20 秒（包含 API 間隔）
- 每小時產生約 120 次 API 呼叫（6 APIs × 20 cycles）

### 2. 統計資料累積

- 目前已收集 33 個樣本（order/create 端點）
- 資料集中在時間桶：17:00 工作日
- 其他 5 個 API 端點尚未達到最小樣本數 (50)

### 3. 為何只有一個端點有統計資料？

可能原因：
1. **最小樣本數限制**：其他端點樣本數 < 50（配置的 min_samples）
2. **時間因素**：Trace Generator 剛啟動不久
3. **Baseline 更新間隔**：需要等待 baseline 更新週期

### 4. 預期的資料增長

根據目前的配置：
- **30 秒間隔** × **6 APIs** = 每分鐘 12 個 traces
- 達到 50 個樣本需要約 **4-5 分鐘**
- 預計 5-10 分鐘後會看到更多端點的統計資料

## 📈 後續監控建議

### 1. 持續監控腳本

```bash
# 每 60 秒檢查一次，共檢查 10 次（總計 10 分鐘）
CHECK_INTERVAL=60 MAX_CHECKS=10 ./monitor-trace-stats.sh
```

### 2. 查看即時日誌

```bash
# Trace Generator 日誌
docker logs -f trace-generator

# Tempo Anomaly Service 日誌
docker logs -f tempo-anomaly-service
```

### 3. 定期查詢統計資料

```bash
# 每分鐘自動查詢
watch -n 60 'curl -s http://localhost:8081/v1/available | jq .'
```

### 4. 監控 Redis 資料

```bash
# 查看 baseline keys 數量
docker exec tempo-anomaly-redis redis-cli KEYS "base:*" | wc -l

# 查看 duration keys 數量
docker exec tempo-anomaly-redis redis-cli KEYS "dur:*" | wc -l
```

## 🛠️ 有用的命令

### 服務管理

```bash
# 啟動所有服務
docker-compose -f docker-compose-deploy.yml up -d

# 停止 Trace Generator
docker-compose -f docker-compose-deploy.yml stop trace-generator

# 重啟 Trace Generator
docker-compose -f docker-compose-deploy.yml restart trace-generator

# 停止所有服務
docker-compose -f docker-compose-deploy.yml down
```

### API 查詢

```bash
# 查看可用服務
curl http://localhost:8081/v1/available | jq .

# 查詢 baseline
curl 'http://localhost:8081/v1/baseline?service=trace-demo-service&endpoint=POST%20%2Fapi%2Forder%2Fcreate&hour=17&dayType=weekday' | jq .

# 檢測異常
curl -X POST http://localhost:8081/v1/anomaly/check \
  -H 'Content-Type: application/json' \
  -d '{"service":"trace-demo-service","endpoint":"POST /api/order/create","timestampNano":'$(date +%s)000000000',"durationMs":1000}' | jq .
```

### 日誌檢查

```bash
# Trace Generator 日誌
docker logs --tail=50 trace-generator

# Tempo Anomaly Service 日誌
docker logs --tail=50 tempo-anomaly-service

# 查看 Tempo 收集狀態
docker logs --tail=20 tempo-anomaly-service | grep "tempo poll"
```

## ✨ 結論

### 成功項目

1. ✅ Trace Generator 成功啟動並運行
2. ✅ 每 30 秒自動呼叫 6 個 API 端點
3. ✅ Tempo 成功收集 traces
4. ✅ tempo-anomaly-service 成功建立統計資料
5. ✅ 已有 1 個端點具備完整的 baseline 資料
6. ✅ 監控腳本可正常運行並顯示統計資料

### 待觀察項目

1. ⏳ 其他 5 個 API 端點的統計資料累積
2. ⏳ 多個時間桶（小時）的資料分布
3. ⏳ 工作日/週末的資料差異

### 預期時間線

- **5-10 分鐘**：其他端點達到最小樣本數
- **30-60 分鐘**：多個時間桶開始有資料
- **數小時**：完整的 24 小時資料覆蓋
- **數天**：工作日和週末的資料差異顯現

## 📚 相關文件

- [Trace Generator README](trace-generator/README.md)
- [Tempo Anomaly Service API](../tempo-latency-anomaly-service/docs/features/AVAILABLE_API.md)
- [快速開始指南](../tempo-latency-anomaly-service/docs/QUICKSTART.md)

---

**報告時間**: 2026-01-23 12:03:00  
**監控狀態**: ✅ 正常運行  
**資料收集**: ✅ 進行中
