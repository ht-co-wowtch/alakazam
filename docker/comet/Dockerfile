FROM alpine

WORKDIR /usr/local/bin

COPY docker-entrypoint.sh .
COPY --from=alakazams_build:latest /comet .

RUN chmod +x docker-entrypoint.sh

ENTRYPOINT ["docker-entrypoint.sh"]

CMD ["comet", "-c", "config.yml"]
