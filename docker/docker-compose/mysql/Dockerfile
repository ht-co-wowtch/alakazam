ARG MYSQL_VERSION=latest
FROM mysql:${MYSQL_VERSION}

# Set timezone
ARG TZ=UTC
ENV TZ ${TZ}
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone && chown -R mysql:root /var/lib/mysql/

# Set UTF8 encoded for initdb
ENV LANG=C.UTF-8

CMD ["mysqld"]

EXPOSE 3306
