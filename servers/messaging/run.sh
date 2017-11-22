#!usr/bin/env bash

set -e

export MESSAGING_CONTAINER=info-344-messaging
export MQ_CONTAINER=rabbitmq-server
export APP_NETWORK=appnet

docker pull zicodeng/$MESSAGING_CONTAINER

if [ "$(docker ps -aq --filter name=$MESSAGING_CONTAINER)" ]; then
    docker rm -f $MESSAGING_CONTAINER
fi

if [ "$(docker images -q -f dangling=true)" ]; then
    docker rmi $(docker images -q -f dangling=true)
fi

docker system prune -f

if ! [ "$(docker network ls | grep $APP_NETWORK)" ]; then
    docker network create $APP_NETWORK
fi

# No need to specify Redis port here,
# because it is default to 6379.
docker run \
-d \
-e ADDR=$MESSAGING_CONTAINER:80 \
-e MQADDR=$MQ_CONTAINER:5672 \
-e DBADDR=mongo-server:27017 \
-e REDISADDR=redis-server \
-e SUMMARYSVCADDR=info-344-summary:80 \
--name $MESSAGING_CONTAINER \
--network $APP_NETWORK \
--restart unless-stopped \
zicodeng/$MESSAGING_CONTAINER
