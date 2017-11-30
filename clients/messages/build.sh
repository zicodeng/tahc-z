#!/usr/bin/env bash

set -e

docker build -t zicodeng/info-344-client .

if [ "$(docker ps -aq --filter name=info-344-client)" ]; then
    docker rm -f info-344-client
fi

# Remove dangling images.
if [ "$(docker images -q -f dangling=true)" ]; then
    docker rmi $(docker images -q -f dangling=true)
fi