
## Build
```
    make build
```

### Run
```
    make run
    make stop

    // or
    go run cmd/logic/main.go -c cmd/logic/logic-example.yml
    go run cmd/comet/main.go -c cmd/comet/comet-example.yml
    go run cmd/job/main.go -c cmd/job/job-example.yml

```

### Dependencies
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

