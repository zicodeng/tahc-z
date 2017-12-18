#!/usr/bin/env bash

set -e

export ADDR=localhost:5000
export REDISADDR=192.168.99.100:6379

go run main.go