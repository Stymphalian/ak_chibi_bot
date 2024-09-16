#!/usr/bin/env bash

cd /root/dev/ak_chibi_bot/db
mkdir -p certs
cp /etc/letsencrypt/live/psql.stymphalian.top/fullchain.pem certs/fullchain.pem
cp /etc/letsencrypt/live/psql.stymphalian.top/privkey.pem certs/privkey.pem
docker build -t ak-db:prod -f Dockerfile --target production .

# cd /root/dev/ak_chibi_bot/
# docker compose -f compose.yaml restart db
