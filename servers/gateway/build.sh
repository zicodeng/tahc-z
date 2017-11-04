#!/usr/bin/env bash

# Build the API server Linux executable.
GOOS=linux go build

# Build the API server Docker container image.
docker build -t zicodeng/info-344-api .

if [ "$(docker ps -aq --filter name=info-344-api)" ]; then
    docker rm -f info-344-api
fi

# This command is not working as expected?
# docker image prune -f

# Remove dangling images.
if [ "$(docker images -q -f dangling=true)" ]; then
    docker rmi $(docker images -q -f dangling=true)
fi

go clean