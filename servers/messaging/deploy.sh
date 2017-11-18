#!/usr/bin/env bash

set -e

export MESSAGING_CONTAINER=info-344-messaging

./build.sh

docker push zicodeng/$MESSAGING_CONTAINER

ssh -oStrictHostKeyChecking=no root@107.170.225.128 'bash -s' < run.sh