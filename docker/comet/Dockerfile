ARG image

FROM ${image} as base
FROM alpine

WORKDIR /usr/local/bin

COPY --from=base /usr/local/bin/comet.yml config.yml
COPY --from=base /usr/local/bin/comet .

CMD ["comet", "-c", "config.yml"]