#!/usr/bin/env bash

docker build -t zicodeng/info-344-client .

if [ "$(docker ps -aq --filter name=info-344-client)" ]; then
    docker rm -f info-344-client
fi