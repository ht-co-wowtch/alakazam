ARG image

FROM ${image} as base
FROM alpine

WORKDIR /usr/local/bin

COPY docker-entrypoint.sh .
COPY --from=base /usr/local/bin/admin .

RUN chmod +x docker-entrypoint.sh

ENTRYPOINT ["docker-entrypoint.sh"]

CMD ["admin", "-c", "config.yml"]