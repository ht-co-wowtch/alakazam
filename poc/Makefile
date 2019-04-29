# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test

build:
	rm -rf bin/
	mkdir bin/
	$(GOBUILD) -o bin/web cmd/web/main.go

run:
	nohup bin/web  2>&1 > bin/web.log &

stop:
	pkill -f bin/web
