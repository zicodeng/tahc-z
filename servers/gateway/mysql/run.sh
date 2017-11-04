#!/usr/bin/env bash
set -e

./build.sh

if [ "$(docker ps -aq --filter name=$CONTAINER_NAME)" ]; then
    docker rm -f $CONTAINER_NAME
fi

docker run -d \
--name $CONTAINER_NAME \
-p 3306:3306 \
-e MYSQL_ROOT_PASSWORD=$MYSQL_ROOT_PASSWORD \
-e MYSQL_DATABASE=$MYSQL_DATABASE \
$INFO_344_MYSQL_IMAGE
