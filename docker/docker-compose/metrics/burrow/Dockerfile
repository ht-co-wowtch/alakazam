FROM golang:1.12-alpine as builder

ENV GO111MODULE=on

RUN apk add --no-cache git

WORKDIR app

COPY burrow.toml /etc/burrow/

RUN git clone https://github.com/linkedin/Burrow.git .

RUN go mod download && go build -ldflags '-s -w' -o /tmp/burrow .

FROM alpine

COPY --from=builder /tmp/burrow /app/
COPY --from=builder /etc/burrow/ /etc/burrow/

CMD ["/app/burrow", "--config-dir", "/etc/burrow"]