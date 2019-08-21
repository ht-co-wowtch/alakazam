ARG image

FROM ${image} as base
FROM alpine

WORKDIR /usr/local/bin

COPY --from=base /usr/local/bin/seq.yml config.yml
COPY --from=base /usr/local/bin/seq .

CMD ["seq", "-c", "config.yml"]