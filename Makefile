# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test

all: test build
build:
	rm -rf bin/
	mkdir bin/
	cp cmd/comet/comet-example.toml bin/comet.toml
	cp cmd/logic/logic-example.toml bin/logic.toml
	cp cmd/job/job-example.toml bin/job.toml
	$(GOBUILD) -o bin/comet cmd/comet/main.go
	$(GOBUILD) -o bin/logic cmd/logic/main.go
	$(GOBUILD) -o bin/job cmd/job/main.go

test:
	$(GOTEST) -v ./...

clean:
	rm -rf bin/

run:
	nohup bin/logic -conf=bin/logic.toml -region=sh -zone=sh001 -deploy.env=dev -weight=10 2>&1 > bin/logic.log &
	nohup bin/comet -conf=bin/comet.toml -region=sh -zone=sh001 -deploy.env=dev -weight=10 -addrs=127.0.0.1 2>&1 > bin/comet.log &
	nohup bin/job -conf=bin/job.toml -region=sh -zone=sh001 -deploy.env=dev 2>&1 > bin/job.log &

stop:
	pkill -f bin/logic
	pkill -f bin/job
	pkill -f bin/comet
