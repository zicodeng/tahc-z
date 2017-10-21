#!/usr/bin/env bash

# Remove the old container first.
docker rm -f info-344-api

docker run -d \
-p 443:443 \
--name info-344-api \
-v "$(pwd)/tls":/tls:ro \
-e TLSCERT=/tls/fullchain.pem \
-e TLSKEY=/tls/privkey.pem \
zicodeng/info-344-api