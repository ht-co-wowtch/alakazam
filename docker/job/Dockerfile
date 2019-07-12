ARG image

FROM ${image} as base
FROM alpine

WORKDIR /usr/local/bin

COPY --from=base /usr/local/bin/job.yml config.yml
COPY --from=base /usr/local/bin/job .


CMD ["job", "-c", "config.yml"]