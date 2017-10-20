#!/usr/bin/env bash

./build.sh

docker push zicodeng/info-344-client

ssh root@107.170.241.115 "./run.sh"