#!/usr/bin/env bash

# This file will be running on the cloud.
# Linux machine expects file with LF line endings instead of CRLF.
# Make sure the file is saved with appropriate line endings.

export API_CONTAINER=info-344-api
export REDIS_CONTAINER=redis-server
export MONGO_CONTAINER=mongo-server
export APP_NETWORK=appnet

export TLSCERT=/etc/letsencrypt/live/info-344-api.zicodeng.me/fullchain.pem
export TLSKEY=/etc/letsencrypt/live/info-344-api.zicodeng.me/privkey.pem

export SESSIONKEY=secretsigningkey

export ADDR=:443
export REDISADDR=$REDIS_CONTAINER:6379
export DBADDR=$MONGO_CONTAINER:27017

# Make sure to get the latest image.
docker pull zicodeng/info-344-api

# Remove the old containers first.
if [ "$(docker ps -aq --filter name=$API_CONTAINER)" ]; then
    docker rm -f $API_CONTAINER
fi

if [ "$(docker ps -aq --filter name=$REDIS_CONTAINER)" ]; then
    docker rm -f $REDIS_CONTAINER
fi

if [ "$(docker ps -aq --filter name=$MONGO_CONTAINER)" ]; then
    docker rm -f $MONGO_CONTAINER
fi

# Remove dangling images.
if [ "$(docker images -q -f dangling=true)" ]; then
    docker rmi $(docker images -q -f dangling=true)
fi

# Create Docker private network if not exist.
if ! [ "$(docker network ls | grep $APP_NETWORK)" ]; then
    docker network create $APP_NETWORK
fi

# Run Redis Docker container inside our appnet private network.
docker run \
-d \
--name $REDIS_CONTAINER \
--network $APP_NETWORK \
--restart unless-stopped \
redis

# Run Mongo Docker container.
docker run \
-d \
--name mongo-server \
--network $APP_NETWORK \
--restart unless-stopped \
mongo

# Run Info 344 API Docker container.
docker run \
-d \
-p 443:443 \
--name info-344-api \
--network $APP_NETWORK \
-v /etc/letsencrypt:/etc/letsencrypt:ro \
-e TLSCERT=$TLSCERT \
-e TLSKEY=$TLSKEY \
-e SESSIONKEY=$SESSIONKEY \
-e ADDR=$ADDR \
-e REDISADDR=$REDISADDR \
-e DBADDR=$DBADDR \
--restart unless-stopped \
zicodeng/info-344-api