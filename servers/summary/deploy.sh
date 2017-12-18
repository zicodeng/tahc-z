#!/usr/bin/env bash

set -e

./build.sh

docker push zicodeng/info-344-summary

ssh -oStrictHostKeyChecking=no root@107.170.225.128 'bash -s' < run.sh