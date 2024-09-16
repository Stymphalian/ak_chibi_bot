# docker build -t ak-db --target development .
# docker build -t ak-db:prod --target production .
FROM postgres:16.4 AS base
COPY ./configs/pg_hba.conf /opt/postgres/pg_hba.conf
COPY ./configs/postgresql.conf /opt/postgres/postgresql.conf
COPY ./configs/50-hba.sh /docker-entrypoint-initdb.d/50-hba.sh
COPY ./configs/100-web_user.sql /docker-entrypoint-initdb.d/100-web_user.sql
RUN chown postgres:postgres /opt/postgres/pg_hba.conf
RUN chown postgres:postgres /opt/postgres/postgresql.conf
RUN chmod 600 /opt/postgres/pg_hba.conf
RUN chmod 600 /opt/postgres/postgresql.conf

ADD ./certs/fullchain.pem /var/lib/postgresql/fullchain.pem
ADD ./certs/privkey.pem /var/lib/postgresql/privkey.pem
RUN chown postgres:postgres /var/lib/postgresql/fullchain.pem
RUN chown postgres:postgres /var/lib/postgresql/privkey.pem
RUN chmod 600 /var/lib/postgresql/fullchain.pem
RUN chmod 600 /var/lib/postgresql/privkey.pem

FROM base AS development
CMD ["postgres", "-c", "config_file=/opt/postgres/postgresql.conf"]

FROM base AS production
CMD ["postgres", "-c", "config_file=/opt/postgres/postgresql.conf"]
