#!/usr/bin/env bash
# Bash script for updating the PSQL certs upon certbot renewal

if [ "$RENEWED_LINEAGE" == "/etc/letsencrypt/live/psql.stymphalian.top" ]; then
    echo "Renewed $RENEWED_LINEAGE"
    echo "Renewed $RENEWED_LINEAGE" >> /root/changed.txt
    /root/dev/ak_chibi_bot/db/update_certs.sh
else
    echo "Renewed $RENEWED_LINEAGE"
    echo "Other" 
    echo "Renewed $RENEWED_LINEAGE" >> /root/not_changed.txt
fi

