#!/usr/bin/env bash

# Build the API server Linux executable and API server Docker container image.
./build.sh

# Push the image to my DockerHub.
docker push zicodeng/info-344-api

# Run run.sh remotely.
ssh root@107.170.225.128 "./run.sh"