.PHONY: all test clean

GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test

REGISTRY_IMAGE ?= alakazam
REGISTRY_TAG ?= latest
REGISTRY_NAME := $(REGISTRY_IMAGE):$(REGISTRY_TAG)

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
	nohup bin/logic -c logic.yml 2>&1 > bin/logic.log &
	nohup bin/comet -c comet.yml 2>&1 > bin/comet.log &
	nohup bin/job -c job.yml 2>&1 > bin/job.log &
	nohup bin/admin -c admin.yml 2>&1 > bin/admin.log &

stop:
	pkill -f bin/comet
	pkill -f bin/admin
	pkill -f bin/logic
	pkill -f bin/job

migrate:
	bin/logic -c logic.yml -migrate=true

proto-build: proto-logic proto-comet

proto-logic:
	cd app/logic/pb && protoc \
	--proto_path=${GOPATH}/src \
	--proto_path=${GOPATH}/src/github.com/gogo/protobuf/protobuf \
	--proto_path=. \
	--gofast_out=plugins=grpc:. *.proto \
	*.proto

proto-comet:
	cd app/comet/pb && protoc \
	--proto_path=${GOPATH}/src \
	--proto_path=${GOPATH}/src/github.com/gogo/protobuf/protobuf \
	--proto_path=. \
	--gofast_out=plugins=grpc:. *.proto \
	*.proto

test:
	sh test/unit

test-cover: test cover

cover:
	if [ -f coverage.out ]; then \
		go tool cover -html coverage.out -o coverage.html; \
	fi

docker-clean:
	docker rmi `docker images -q --filter 'reference=alakaza*'`

docker-clean-service:
	docker rmi `docker images -q --filter 'reference=alakazam_*'`

docker-build:
	docker build -t $(REGISTRY_NAME) .

docker-build-service:
	cd docker && sh build.sh

docker-run:
	docker run --name alakazam_logic -it -d -v logic.yml:/logic.yml alakazam_logic
