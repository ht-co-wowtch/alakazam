### Build
```
    make build
```

### Run
```
    make run
    make stop

    // or
    go run cmd/logic/main.go -conf=bin/logic.toml -region=sh -zone=sh001 -deploy.env=dev -weight=10 2>&1 > bin/logic.log &
    go run cmd/comet/main.go -conf=bin/comet.toml -region=sh -zone=sh001 -deploy.env=dev -weight=10 -addrs=127.0.0.1 2>&1 > bin/logic.log &
    go run cmd/job/main.go -conf=bin/job.toml -region=sh -zone=sh001 -deploy.env=dev 2>&1 > bin/logic.log &

```

### Dependencies
[Discovery](https://github.com/bilibili/discovery)

[Kafka](https://kafka.apache.org/quickstart)