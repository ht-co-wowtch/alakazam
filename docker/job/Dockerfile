ARG image

FROM ${image} as base
FROM alpine

WORKDIR /usr/local/bin

COPY docker-entrypoint.sh .
COPY --from=base /usr/local/bin/job .

RUN chmod +x docker-entrypoint.sh

ENTRYPOINT ["docker-entrypoint.sh"]

CMD ["job", "-c", "config.yml"]