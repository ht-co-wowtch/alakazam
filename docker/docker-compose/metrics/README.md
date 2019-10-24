# 聊天室監控

1. 啟動聊天室

   [docker-compose](https://gitlab.com/jetfueltw/cpw/alakazam/tree/develop/docker/docker-compose#docker-compose)

2. 先copy prometheus-example.yml 

```bash
> cp prometheus-example.yml prometheus.yml
```

3. 檢查各服務狀況

```bash
> curl 127.0.0.1:3030/metrics #127.0.0.1 改填成logic ip
> curl 127.0.0.1:3031/metrics #127.0.0.1 改填成comet ip
> curl 127.0.0.1:3032/metrics #127.0.0.1 改填成job ip
> curl 127.0.0.1:3033/metrics #127.0.0.1 改填成message ip
> curl 127.0.0.1:3034/metrics #127.0.0.1 改填成seq ip
> curl 127.0.0.1:3035/metrics #127.0.0.1 改填成admin ip
> curl 127.0.0.1:3036/metrics #127.0.0.1 改填成kafka ip

# 只要確認會回傳一些 text/plain 即可，沒有出現500 or 404
```

4. 設定prometheus.yml內所有job targets

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
```

5. docker run 

```bash
> docker-compose up -d 
```

6. prometheus  port `9090 ` and  grafana port `3000`都對外開放並限制特定IP可存取，ex 台南 and 台中辦公室ip

7. 打開`127.0.0.1:9090/targets`瀏覽器確認prometheus監控狀況，`127.0.0.1`自行更改成prometheus主機上的ip，確認是否都為Status為`UP`

   ![arch](./doc/prometheus_status.png)

8. 打開`127.0.0.1:3000`瀏覽器確認grafana狀況 

![arch](./doc/grafana.png)

7. dashboards 目錄有多個監控dashboard可以匯入`grafana`