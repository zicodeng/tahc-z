#!/usr/bin/env bash

# Build the API server Linux executable.
GOOS=linux go build

# Build the API server Docker container image.
docker build -t zicodeng/info-344-api .

docker rm -f info-344-api

# This command is not working as expected?
# docker image prune -f

docker rmi $(docker images -q -f dangling=true)

go clean
