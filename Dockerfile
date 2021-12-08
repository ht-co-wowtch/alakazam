FROM golang:1.13-alpine AS build_module

ENV GO111MODULE=on
ARG GOPROXY
ENV GOPROXY=${GOPROXY}
ENV GOPRIVATE=gitlab.com/jetfueltw/cpw

WORKDIR $GOPATH/src/gitlab.com/jetfueltw/cpw/alakazam

COPY go.mod go.sum ./

RUN apk add --no-cache git && \
    git config --global url."https://gitlab+deploy-token-678908:s1jR1Pt-yvNHrC_9expc@gitlab.com/ht-co/cpw/micro".insteadOf "https://gitlab.com/ht-co/cpw/micro" && \
    go mod download

FROM build_module AS build

ENV GO111MODULE=on
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

COPY . .

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
