FROM golang:1.13 AS build

ADD . /opt/app
WORKDIR /opt/app
RUN go build .


FROM ubuntu:18.04 AS release

ENV PGVER 10
RUN apt -y update && apt install -y postgresql-$PGVER

USER postgres

ADD ./db.sql /opt/db.sql
RUN /etc/init.d/postgresql start &&\
	psql --command "CREATE USER subd WITH SUPERUSER PASSWORD 'subd';" &&\
	createdb -O subd subd &&\
    psql -f /opt/db.sql -d subd &&\
    /etc/init.d/postgresql stop
ENV POSTGRES_DSN=postgres://subd:subd@localhost/subd

# Adjust PostgreSQL configuration so that remote connections to the
# database are possible.
RUN echo "host all  all    0.0.0.0/0  md5" >> /etc/postgresql/$PGVER/main/pg_hba.conf
RUN echo "listen_addresses='*'" >> /etc/postgresql/$PGVER/main/postgresql.conf

EXPOSE 5432

VOLUME ["/etc/postgresql", "/var/log/postgresql", "/var/lib/postgresql"]

USER root

EXPOSE 5000

COPY --from=build /opt/app/technopark-sem2-db /usr/bin/

CMD service postgresql start && technopark-sem2-db