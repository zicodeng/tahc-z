#!/usr/bin/env bash

set -e

# Build the API server Linux executable and API server Docker container image.
./build.sh

# Push the image to my DockerHub.
docker push zicodeng/info-344-gateway

# Send run.sh to the cloud running remotely.
ssh -oStrictHostKeyChecking=no root@107.170.225.128 'bash -s' < run.sh