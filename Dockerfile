FROM golang:1.12-alpine AS build_module

ENV GO111MODULE=on

RUN set -ex && apk add --no-cache git

WORKDIR $GOPATH/src/gitlab.com/jetfueltw/cpw/alakazam

COPY go.mod .
COPY go.sum .

RUN git config --global url."https://gitlab+deploy-token-78991:HcgtGG-9MnBasMsTvNfo@gitlab.com/jetfueltw/cpw/micro".insteadOf "https://gitlab.com/jetfueltw/cpw/micro" && \
    go mod download

FROM build_module AS build

ENV GO111MODULE=on
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

COPY . .

RUN set -ex

RUN go build -o /logic cmd/logic/main.go && \
    go build -o /comet cmd/comet/main.go && \
    go build -o /job cmd/job/main.go && \
    go build -o /admin cmd/admin/main.go