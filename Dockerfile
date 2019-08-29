FROM golang:1.12-alpine AS build_module

ENV GO111MODULE=on
ARG GOPROXY
ENV GOPROXY=${GOPROXY}

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

RUN go build -ldflags '-s -w' -o /logic cmd/logic/main.go && \
    go build -ldflags '-s -w' -o /comet cmd/comet/main.go && \
    go build -ldflags '-s -w' -o /job cmd/job/main.go && \
    go build -ldflags '-s -w' -o /admin cmd/admin/main.go && \
    go build -ldflags '-s -w' -o /message cmd/message/main.go && \
    go build -ldflags '-s -w' -o /seq cmd/seq/main.go

FROM alpine

WORKDIR /usr/local/bin

COPY config/admin-example.yml admin.yml
COPY config/comet-example.yml comet.yml
COPY config/job-example.yml job.yml
COPY config/logic-example.yml logic.yml
COPY config/message-example.yml message.yml
COPY config/seq-example.yml seq.yml
COPY --from=build /logic .
COPY --from=build /comet .
COPY --from=build /job .
COPY --from=build /admin .
COPY --from=build /message .
COPY --from=build /seq .

COPY --from=build /usr/local/go/lib/time/zoneinfo.zip /usr/local/go/lib/time/zoneinfo.zip

CMD ["/bin/sh"]