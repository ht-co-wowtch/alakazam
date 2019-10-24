# docker-compose
整合 alakazam 的環境與服務。

## 環境需求

docker and docker-compose



## 開始使用

git clone 到任何你喜歡的位子後，先 `cp .env.example .env` 再更改你需要的設定，例如：
```
MYSQL_ROOT_PASS=root
```



## 登入 GitLab Container Registry

先到 Gitlab 產生一個 [Personal Access Tokens](https://gitlab.com/profile/personal_access_tokens)，name 填 Container Registry（或其他好記的名字），scopes 勾選 read_registry，建立後記得把 token 記起來。  
之後 terminal 輸入 `docker login -u yourname@cqcp.com.tw registry.gitlab.com`，密碼則是剛才產生的 token。



## 啟動與停止服務

請指定你要啟動的服務，不然會全部啟動，可以配合 alias 節省打字時間。
```
// 啟動 alakazam，他會自動把依賴的服務也跑起來
docker-compose up -d alakazam

// 停止所有服務
docker-compose down
```



## Database Migration

絕大多數的服務都依賴於資料庫，與正確的 schema 版本，在開始開發前，你需要先初始化資料庫。  
第一次會預設建立 `platform`、`alakazam` 兩個資料庫，如果你在 .env 使用其他名字的話，你必須手動新增資料庫再跑 migrate。
```
// 撤銷所有 migration
make platform.rollback

// 跑還沒跑過的 migration
make platform.migrate

// 塞入預設的必要與測試資料
make platform.seed

// 重設整個資料庫，等同於：rollback + migration + seed
make platform.reset
```



## MySQL initdb

當第一次啟動 MySQL 時，會執行 `docker-entrypoint-initdb.d` 資料夾底下的 `.sh`、`.sql` 與 `.sql.gz`，你可以在裡面放初始化資料庫的語法。第一次啟動會跑一段時間才能訪問，大約 1 分鐘左右。  
如果你想刪除所有資料庫並重跑 init，可以刪除 volume 後再啟動。

```
docker-compose down
rm -rf ./data/mysql
docker-compose up -d mysql
```



## Metrics

1. 先copy prometheus-example.yml 

```bash
> cp prometheus-example.yml prometheus.yml
```

2. 檢查各服務狀況

```bash
> curl 127.0.0.1:3030/metrics #127.0.0.1 改填成logic ip
> curl 127.0.0.1:3031/metrics #127.0.0.1 改填成comet ip
> curl 127.0.0.1:3032/metrics #127.0.0.1 改填成job ip
> curl 127.0.0.1:3033/metrics #127.0.0.1 改填成message ip
> curl 127.0.0.1:3034/metrics #127.0.0.1 改填成seq ip
> curl 127.0.0.1:3035/metrics #127.0.0.1 改填成admin ip
> curl 127.0.0.1:3036/metrics #127.0.0.1 改填成kafka ip
> curl 127.0.0.1:3037/metrics #127.0.0.1 改填成burrow ip

# 只要確認會回傳一些 text/plain 即可，沒有出現500 or 404
```

3. 設定prometheus.yml內所有job targets

```yml
scrape_configs:
  - job_name: 'logic'  
    static_configs:
      - targets: ['127.0.0.1:3030']   #127.0.0.1 改填成logic ip

  - job_name: 'comet'
    static_configs:
      - targets: ['127.0.0.1:3031']   #127.0.0.1 改填成comet ip

  - job_name: 'job'
    static_configs:
      - targets: ['127.0.0.1:3032']   #127.0.0.1 改填成job ip

  - job_name: 'message'
    static_configs:
      - targets: ['127.0.0.1:3033']   #127.0.0.1 改填成message ip

  - job_name: 'seq'
    static_configs:
      - targets: ['127.0.0.1:3034']   #127.0.0.1 改填成seq ip

  - job_name: 'admin'
    static_configs:
      - targets: ['127.0.0.1:3035']   #127.0.0.1 改填成admin ip
  
  - job_name: 'kafka_broker_jmx'
    static_configs:
      - targets: ['127.0.0.1:3036']   #127.0.0.1 改填成kafka ip
      
  - job_name: 'kafka_consumer_lag'
    static_configs:
      - targets: ['127.0.0.1:3036']   #127.0.0.1 改填成burrow ip
```

4. docker run 

```bash
> docker-compose up -d 
```

5. prometheus  port `9090 ` and  grafana port `3000`都對外開放並限制特定IP可存取，ex 台南 and 台中辦公室ip

6. 打開`127.0.0.1:9090/targets`瀏覽器確認prometheus監控狀況，`127.0.0.1`自行更改成prometheus主機上的ip，確認是否都為Status為`UP`

   ![arch](./metrics/doc/prometheus_status.png)

7. 打開`127.0.0.1:3000`瀏覽器確認grafana狀況 

![arch](./metrics/doc/grafana.png)

8. dashboards 目錄有多個監控dashboard可以匯入`grafana`