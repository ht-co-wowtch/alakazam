ARG image

FROM ${image} as base
FROM alpine

WORKDIR /usr/local/bin

COPY --from=base /usr/local/go/lib/time/zoneinfo.zip /usr/local/go/lib/time/zoneinfo.zip
COPY --from=base /usr/local/bin/message.yml config.yml
COPY --from=base /usr/local/bin/message .

CMD ["message", "-c", "config.yml"]