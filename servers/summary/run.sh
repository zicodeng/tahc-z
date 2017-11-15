#!usr/bin/env bash

set -e

export ADDR=info-344-summary:80
export REDISADDR=redis-server:6379

export SUMMARY_CONTAINER=info-344-summary
export APP_NETWORK=appnet

docker pull zicodeng/$SUMMARY_CONTAINER

if [ "$(docker ps -aq --filter name=$SUMMARY_CONTAINER)" ]; then
    docker rm -f $SUMMARY_CONTAINER
fi

if [ "$(docker images -q -f dangling=true)" ]; then
    docker rmi $(docker images -q -f dangling=true)
fi

docker system prune -f

if ! [ "$(docker network ls | grep $APP_NETWORK)" ]; then
    docker network create $APP_NETWORK
fi

docker run \
-d \
-e ADDR=$ADDR \
-e REDISADDR=$REDISADDR \
--name $SUMMARY_CONTAINER \
--network $APP_NETWORK \
--restart unless-stopped \
zicodeng/$SUMMARY_CONTAINER
