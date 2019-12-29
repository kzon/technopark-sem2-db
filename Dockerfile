FROM ubuntu:18.04

EXPOSE 5432
EXPOSE 5000

ENV DEBIAN_FRONTEND 'noninteractive'
ENV PGVER 10
RUN apt -y update && apt install -y postgresql-$PGVER
RUN apt install -y wget
RUN apt install -y git

USER postgres

ADD ./db.sql /opt/db.sql
RUN /etc/init.d/postgresql start &&\
	psql --command "CREATE USER subd WITH SUPERUSER PASSWORD 'subd';" &&\
	createdb -O subd subd &&\
    psql -f /opt/db.sql -d subd &&\
    /etc/init.d/postgresql stop
ENV POSTGRES_DSN=postgres://subd:subd@localhost/subd

RUN echo "host all all 0.0.0.0/0 md5" >> /etc/postgresql/$PGVER/main/pg_hba.conf
RUN echo "include_dir='conf.d'" >> /etc/postgresql/$PGVER/main/postgresql.conf
ADD ./postgresql.conf /etc/postgresql/$PGVER/main/conf.d/basic.conf

VOLUME ["/etc/postgresql", "/var/log/postgresql", "/var/lib/postgresql"]


USER root

RUN wget https://dl.google.com/go/go1.13.linux-amd64.tar.gz
RUN tar -xvf go1.13.linux-amd64.tar.gz
RUN mv go /usr/local

ENV GOROOT /usr/local/go
ENV GOPATH /opt/go
ENV PATH $GOROOT/bin:$GOPATH/bin:/usr/local/go/bin:$PATH

ADD . /opt/app
WORKDIR /opt/app
CMD service postgresql start && go run .