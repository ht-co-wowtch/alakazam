.PHONY: all test clean

GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test

build: clean
	$(GOBUILD) -o bin/comet cmd/comet/main.go
	$(GOBUILD) -o bin/logic cmd/logic/main.go
	$(GOBUILD) -o bin/job cmd/job/main.go
	$(GOBUILD) -o bin/admin cmd/admin/main.go

build-debug: clean
	$(GOBUILD) -gcflags "all=-N -l" -o bin/comet cmd/comet/main.go
	$(GOBUILD) -gcflags "all=-N -l" -o bin/logic cmd/logic/main.go
	$(GOBUILD) -gcflags "all=-N -l" -o bin/job cmd/job/main.go
	$(GOBUILD) -gcflags "all=-N -l" -o bin/admin cmd/admin/main.go

clean:
	rm -rf bin/
	mkdir bin/

run:
	nohup bin/logic -c logic.yml -stderrthreshold=INFO 2>&1 > bin/logic.log &
	nohup bin/comet -c comet.yml -stderrthreshold=INFO 2>&1 > bin/comet.log &
	nohup bin/job -c job.yml -stderrthreshold=INFO 2>&1 > bin/job.log &
	nohup bin/admin -c admin.yml 2>&1 > bin/admin.log &

stop:
	pkill -f bin/logic
	pkill -f bin/job
	pkill -f bin/comet
	pkill -f bin/admin

migrate:
	bin/logic -c bin/logic.yml -migrate=true

proto-build:
	cd protocol/proto && protoc \
	--proto_path=${GOPATH}/src \
	--proto_path=${GOPATH}/src/github.com/gogo/protobuf/protobuf \
	--proto_path=. \
	--gofast_out=plugins=grpc:../grpc *.proto \
	*.proto

test:
	sh test/unit

test-cover: test cover

cover:
	if [ -f coverage.out ]; then \
		go tool cover -html coverage.out -o coverage.html; \
	fi

