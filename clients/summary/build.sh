#!/usr/bin/env bash

docker build -t zicodeng/info-344-client .

docker rmi $(docker images -q -f dangling=true)