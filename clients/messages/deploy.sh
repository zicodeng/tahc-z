#!/usr/bin/env bash

set -e

./build.sh

docker push zicodeng/info-344-client

ssh -oStrictHostKeyChecking=no root@107.170.241.115 'bash -s' < run.sh