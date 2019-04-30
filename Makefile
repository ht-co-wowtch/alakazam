# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test

all: test build
build:
	rm -rf bin/
	mkdir bin/
	cp cmd/comet/comet-example.yml bin/comet.yml
	cp cmd/logic/logic-example.yml bin/logic.yml
	cp cmd/job/job-example.yml bin/job.yml
	$(GOBUILD) -o bin/comet cmd/comet/main.go
	$(GOBUILD) -o bin/logic cmd/logic/main.go
	$(GOBUILD) -o bin/job cmd/job/main.go

test:
	$(GOTEST) -v ./...

clean:
	rm -rf bin/

run:
	nohup bin/logic -conf=bin/logic.yml 2>&1 > bin/logic.log &
	nohup bin/comet -conf=bin/comet.yml 2>&1 > bin/comet.log &
	nohup bin/job -conf=bin/job.yml 2>&1 > bin/job.log &

stop:
	pkill -f bin/logic
	pkill -f bin/job
	pkill -f bin/comet
