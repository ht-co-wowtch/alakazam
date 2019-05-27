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
	$(GOBUILD) -o bin/migration cmd/migration/main.go

test:
	$(GOTEST) -v ./...

clean:
	rm -rf bin/

run:
	nohup bin/logic -c bin/logic.yml -stderrthreshold=INFO 2>&1 > bin/logic.log &
	nohup bin/comet -c bin/comet.yml -stderrthreshold=INFO 2>&1 > bin/comet.log &
	nohup bin/job -c bin/job.yml -stderrthreshold=INFO 2>&1 > bin/job.log &

migration:
	./bin/migration -run -c ./bin/logic.yml

stop:
	pkill -f bin/logic
	pkill -f bin/job
	pkill -f bin/comet

proto-build:
	cd protocol/proto && protoc \
	--proto_path=${GOPATH}/src \
	--proto_path=${GOPATH}/src/github.com/gogo/protobuf/protobuf \
	--proto_path=. \
	--gofast_out=plugins=grpc:../grpc *.proto \
	*.proto


