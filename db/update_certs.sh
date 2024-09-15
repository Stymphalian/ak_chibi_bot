#!/usr/bin/env bash

cd /root/dev/ak_chibi_bot/db
mkdir -p certs
cp /etc/letsencrypt/live/psql.jordanyu.com/fullchain.pem certs/fullchain.pem
cp /etc/letsencrypt/live/psql.jordanyu.com/privkey.pem certs/privkey.pem
docker build -t ak-db -f Dockerfile.PROD .

# cd /root/dev/ak_chibi_bot/
# docker compose -f compose.yaml restart db
