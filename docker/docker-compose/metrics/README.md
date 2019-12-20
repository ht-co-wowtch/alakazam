# metrics

整合 alakazam 的環境與服務監控

- [目錄解說](目錄解說)
- [prometheus](#prometheus)
- [kafka](#kafka)
- [burrow](#burrow)
- [grafana](#grafana)
- [總結](#總結)



## 目錄解說

```bash
├── README.md
├── dashboards
├── doc
├── grafana
├── prometheus
└── prometheus-example.yml
```

`dashboards` : grafana的dashboard資料

`grafana` : grafana docker資料位置

`prometheus` : prometheus docker資料位置



## prometheus

1. 先copy prometheus-example.yml 

```bash
cp prometheus-example.yml prometheus.yml
```

2. 設定prometheus.yml內所有job targets，請自行更改targets指向各個服務

```yml
scrape_configs:
  - job_name: 'logic'
    static_configs:
      - targets: ['alakazam_logic:3030']

  - job_name: 'comet'
    static_configs:
      - targets: ['alakazam:3031']

  - job_name: 'job'
    static_configs:
      - targets: ['alakazam_job:3032']

  - job_name: 'message'
    static_configs:
      - targets: ['alakazam_message:3033']

  - job_name: 'seq'
    static_configs:
      - targets: ['alakazam_seq:3034']

  - job_name: 'admin'
    static_configs:
      - targets: ['alakazam_admin:3035']

  - job_name: 'kafka_broker_jmx'
    static_configs:
      - targets: ['kafka:3036']

  - job_name: 'kafka_consumer_lag'
    static_configs:
      - targets: ['burrow_prometheus:3037']
```

解釋一下各job對應的服務

1. 聊天室服務

    | Job name | [Gitlab Container Registry](https://gitlab.com/jetfueltw/cpw/alakazam/container_registry) |
    | -------- | ------------------------------------------------------------ |
    | logic    | registry.gitlab.com/jetfueltw/cpw/alakazam/logic             |
    | comet    | registry.gitlab.com/jetfueltw/cpw/alakazam/comet             |
    | job      | registry.gitlab.com/jetfueltw/cpw/alakazam/job               |
    | message  | registry.gitlab.com/jetfueltw/cpw/alakazam/message           |
    | seq      | registry.gitlab.com/jetfueltw/cpw/alakazam/seq               |
    | admin    | registry.gitlab.com/jetfueltw/cpw/alakazam/admin             |

2. Kafka，這部分對應的服務要使用docker部署後續會提到

   | Job name           | 說明                                                         |
   | ------------------ | ------------------------------------------------------------ |
   | kafka_broker_jmx   | 利用jmx方式拿到kafka metrics                                 |
   | kafka_consumer_lag | 利用[Burrow](https://github.com/linkedin/Burrow)監控kafka consumer message lag狀態 |
   
3. 各個服務metrics port皆有預設值，如想要更改則需至對應config做修改

| 服務               | metrics port                                                 |
| ------------------ | ------------------------------------------------------------ |
| logic              | https://gitlab.com/jetfueltw/cpw/alakazam/blob/develop/config/logic-example.yml#L89 |
| comet              | https://gitlab.com/jetfueltw/cpw/alakazam/blob/develop/config/comet-example.yml#L39 |
| job                | https://gitlab.com/jetfueltw/cpw/alakazam/blob/develop/config/job-example.yml#L26 |
| message            | https://gitlab.com/jetfueltw/cpw/alakazam/blob/develop/config/message-example.yml#L51 |
| seq                | https://gitlab.com/jetfueltw/cpw/alakazam/blob/develop/config/seq-example.yml#L48 |
| admin              | https://gitlab.com/jetfueltw/cpw/alakazam/blob/develop/config/admin-example.yml#L70 |
| Kafka broker       | https://gitlab.com/jetfueltw/cpw/alakazam/blob/develop/docker/docker-compose/docker-compose.yml#L250 |
| kafka consumer lag | https://gitlab.com/jetfueltw/cpw/alakazam/blob/develop/docker/docker-compose/docker-compose.yml#L330 |

4. 聊天室本身的服務只要架設好，prometheus.yml設定好就可以拿到metrics

5. kafka metrics則需要另外安裝相對應服務才能獲取metrics，請參考[kafka](#kafka)

   

## kafka

想要監控kafka除了利用一些[client](https://cwiki.apache.org/confluence/display/KAFKA/Clients)取拉資料外，但kafka本身有些內部metrics是不對外開放給client取得，但kafka本身有利用jmx紀錄metrics，於是在kafka啟動時告知要增加jmx功能即可從外部取得kafka metrics

kafka配置jmx方式如下

1. 首先docker需要增加對應的env，[KAFKA_JMX_OPTS](https://github.com/apache/kafka/blob/trunk/bin/kafka-run-class.sh#L182)是告知kafka啟動需要jmx，`Dcom.sun.management.jmxremote.rmi.port`是利用java rmi綁定一個port讓其可以對外讀取，`Djava.rmi.server.hostname`同樣是java rmi綁定一個可以讓外部讀取jmx資料的host，請設定為kafka Container ip

   ```yml
   KAFKA_JMX_OPTS: -Dcom.sun.management.jmxremote -Dcom.sun.management.jmxremote.authenticate=false -Dcom.sun.management.jmxremote.ssl=false -Dcom.sun.management.jmxremote.rmi.port=9091 -Djava.rmi.server.hostname=${KAFKA_JMX_SERVER_HOST}
   ```

   

2. jmx設定完成後，由於原始方式透過java撰寫client端程式去跟kafka jmx取資料或是透過`jconsole`GUI方式觀看，但這兩種都很麻煩，期望的是能將jmx metrics放置在`grafana`，所以利用`prometheus`提供的[jmx_exporter](https://github.com/prometheus/jmx_exporter)收集kafka jmx metrics．利用java javaagent以代理方式將`jmx_exporter`混入kafka中就可以開始讀取kafka jmx metrics．在docker上增加[EXTRA_ARGS](https://github.com/apache/kafka/blob/trunk/bin/kafka-server-start.sh#L44) env可以在啟動kafka時額外指定參數

   ```yml
   EXTRA_ARGS: "-javaagent:/opt/kafka/jmx/jmx_prometheus_javaagent-0.12.0.jar=3036:/opt/kafka/jmx/config.yaml"
   ```

   上述指令可以看到kafka啟動時需一並指定`jmx_exporter.jar`，而kafka是包在docker內執行，所以`jmx_exporter`相關檔案也需要一並放入docker中一起執行，關於`jmx_exporter`則有一個`jar`與`config.yaml`位置在[這](https://gitlab.com/jetfueltw/cpw/alakazam/tree/develop/docker/docker-compose/kafka/jmx)，`3036`就是kafka jmx的port

   

3. 利用`Dcom.sun.management.jmxremote.rmi.port`將一個port設定成可以外部讀取，但此port還未設定給哪個服務使用，此時在docker上增加`JMX_PORT`用以指定該port給kafka jmx使用

   ```yml
   JMX_PORT: 9091
   ```

   

4. 上述總結範例如下，volumes內`kafka/jmx`目錄放著`jmx_prometheus`相關檔案，更詳細範例請[看](https://gitlab.com/jetfueltw/cpw/alakazam/blob/develop/docker/docker-compose/docker-compose.yml#L222)

   ```yml
   kafka:
       image: wurstmeister/kafka:2.12-2.3.0
       environment:
         KAFKA_JMX_OPTS: -Dcom.sun.management.jmxremote -Dcom.sun.management.jmxremote.authenticate=false -Dcom.sun.management.jmxremote.ssl=false -Dcom.sun.management.jmxremote.rmi.port=9091 -Djava.rmi.server.hostname=${KAFKA_JMX_SERVER_HOST}
         EXTRA_ARGS: "-javaagent:/opt/kafka/jmx/jmx_prometheus_javaagent-0.12.0.jar=3036:/opt/kafka/jmx/config.yaml"
         JMX_PORT: 9091
       volumes:
         - /var/run/docker.sock:/var/run/docker.sock
         - ./kafka/jmx:/opt/kafka/jmx
         - ./kafka/data:/var/lib/kafka
       ports:
         - "9091:9091"
         - "3036:3036"
       depends_on:
         - zookeeper
   ```

   

## burrow

[burrow](https://github.com/linkedin/Burrow)是一種監控Kafka Consumer Lag Checking，用來確認Consumer延遲消費訊息太久或是某段時間都無工作，意思就是還有多少筆訊息應該要消費但延遲太久沒有消費，在這一點kafka jmx也是有對應的[kafka.consumer:type=ConsumerFetcherManager,name=MaxLag](https://docs.confluent.io/current/kafka/monitoring.html#old-consumer-metrics)可以做判斷，但是[本非準確](https://github.com/linkedin/Burrow/wiki#why-not-maxlag)，而`burrow`是建立模擬一個kafka Consumer 實際去跟kafka broker 實際做消費然後對比`__consumer_offsets`topic offset計算出來的lag值，會更貼近實際狀況

首先要部署一個`burrow` Container，設定相關參數值如下，`zookeeper`與`kafka`請根據實際服務做設定

```yml
  burrow:
    image: registry.gitlab.com/jetfueltw/cpw/kafka-lag/burrow
    environment:
      - BURROW_ZOOKEEPER_SERVERS=zookeeper:2181
      - BURROW_CLUSTER_LOCAL_SERVERS=kafka:9092
      - BURROW_CONSUMER_LOCAL_SERVERS=kafka:9092
      - BURROW_CONSUMER_LOCAL_BACKFILL_EARLIEST=false
      - BURROW_CONSUMER_LOCAL_START_LATEST=true
    ports:
      - "3400:8080"
```

| env                                | 介紹                                  | 值             |
| ---------------------------------- | ------------------------------------- | -------------- |
| BURROW_ZOOKEEPER_SERVERS           | zookeeper connection host             | zookeeper:2181 |
| BURROW_CLUSTER_LOCAL_SERVERS       | kafka cluster connection broker host  | kafka:9092     |
| BURROW_CONSUMER_LOCAL_SERVERS      | kafka consumer connection broker host | kafka:9092     |
| BURROW_CONSUMER_LOCAL_START_LATEST | Kafka auto.offset.reset設定           | true           |



由於監控UI是`grafana`，所以要透過`prometheus`收集metrics就需要再部署一個`burrow`專用的`prometheus`，如下範例，目前版本是[v1.0.0](https://gitlab.com/jetfueltw/cpw/kafka-lag/container_registry)

```yml
  burrow_prometheus:
    image: registry.gitlab.com/jetfueltw/cpw/kafka-lag/prometheus:${BURROW_PROMETHEUS_VERSION}
    environment:
      - BURROW_PROMETHEUS_ADDR=${PROMETHEUS_ADDR}
      - BURROW_PROMETHEUS_BURROW_ADDR=${BURROW_ADDR}
      - BURROW_PROMETHEUS_INTERVAL=${PROMETHEUS_INTERVAL}
    ports:
      - "3037:3037"
    depends_on:
      - burrow
      - prometheus
```

| env                       | 介紹                        | 值                 |
| ------------------------- | --------------------------- | ------------------ |
| PROMETHEUS_ADDR           | 該prometheus metrics port   | :3037              |
| BURROW_ADDR               | burrow metrics api endpoint | http://burrow:8080 |
| PROMETHEUS_INTERVAL       | prometheus metrics頻率      | 30s                |
| BURROW_PROMETHEUS_VERSION | burrow_prometheus版本       | 1.0.0              |



上述實際應用範例[參考](https://gitlab.com/jetfueltw/cpw/alakazam/blob/develop/docker/docker-compose/docker-compose.yml#L316)



## grafana

將[dashboards](https://gitlab.com/jetfueltw/cpw/alakazam/tree/develop/docker/docker-compose/metrics/dashboards)匯入`grafana`，以下是對應說明

| json檔                     | 說明                                                         |
| -------------------------- | ------------------------------------------------------------ |
| jvm.json                   | kafka運行的jvm metrics，透過jmx取得                          |
| kafka_broker.json          | kafka metrics，透過jmx取得                                   |
| kafka_consumer(job).json   | job服務內有關於kafka consumer metrics，透過job服務取得       |
| kafka_producer(logic).json | logic服務內有關於kafka producer metrics，透過logic服務取得   |
| metrics.json               | 聊天室各[服務](https://gitlab.com/jetfueltw/cpw/alakazam/container_registry)基本的metrics，不含kafka、zk等外部依賴服務 |



## 總結

簡略介紹部署步驟

1. 設定完kafka jmx，如[說明](#kafka)
2. 設定完burrow，如[說明](#burrow)
3. 啟動聊天室各[服務](https://gitlab.com/jetfueltw/cpw/alakazam/container_registry)
4. 啟動`prometheus`
5. 確認`prometheus` status下各服務狀態是否為`UP`
6. 匯入`grafana` dashboards

如上述不知道怎部署，請參考[這裡](https://gitlab.com/jetfueltw/cpw/alakazam/tree/develop/docker/docker-compose)，將整個聊天室運作起來實際操作一次

 
