ARG image

FROM ${image} as base
FROM alpine

WORKDIR /usr/local/bin

COPY docker-entrypoint.sh .
COPY --from=base /usr/local/bin/comet .

RUN chmod +x docker-entrypoint.sh

ENTRYPOINT ["docker-entrypoint.sh"]

CMD ["comet", "-c", "config.yml"]