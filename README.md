# 聊天室
## Build
```
    make build
```

### Run
```
    make run
    make stop
```

### Dependencies
[Redis](#)

[Kafka](https://kafka.apache.org/quickstart)

    cd kafka
    docker-compose up -d
    
### tag

tag|說明|
---|----|
v0.1.0|config改為yml
v0.1.1|移除Discovery
v0.1.2|module更名
v0.1.3|module移除沒用到的package
v0.2.0|更改protobuf目錄結構
v0.3.0|protobuf移除不必要參數與method
v0.4.0|移除有關於Operation推送限制
v0.5.0|移除room type推送限制

## Protocol Body格式

採用websocket做binary傳輸，聊天室推給client訊息內容如下格式

name | length | remork |說明
-----|--------|--------|-----
Package|4 bytes|header + body length| 整個Protocol bytes長度
Header|2 bytes|protocol header length| Package  - Boyd 
Operation|4 bytes| [Operation](#operation)|Protocol的動作
Body |不固定|傳送的資料16bytes之後就是Body

### Package
用於表示本次傳輸的binary內容總長度是多少(header + body)

### Header
用來說明binary Header是多少

### Operation
不同的Operation說明本次Protocol資料是什麼，如心跳回覆,訊息等等

### Body
聊天室的訊息內容

## Operation
用來標示該Protocol的動作

name | 說明 |
-----|-----|
1|要求連線到某一個房間
2|連線到某一個房間結果回覆
3|發送心跳
4|回覆心跳結果
5|聊天室訊息
6|更換房間
7|回覆更換房間結果
