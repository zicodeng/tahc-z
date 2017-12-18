#!/usr/bin/env bash

set -e

# This file will be running on the cloud.
# Linux machine expects file with LF line endings instead of CRLF.
# Make sure the file is saved with appropriate line endings.

export GATEWAY_CONTAINER=info-344-gateway
export REDIS_CONTAINER=redis-server
export MONGO_CONTAINER=mongo-server
export MQ_CONTAINER=rabbitmq-server

export APP_NETWORK=appnet
export DBNAME="info_344"

export TLSCERT=/etc/letsencrypt/live/info-344-api.zicodeng.me/fullchain.pem
export TLSKEY=/etc/letsencrypt/live/info-344-api.zicodeng.me/privkey.pem

export SESSIONKEY=secretsigningkey

export ADDR=:443
export REDISADDR=$REDIS_CONTAINER:6379
export DBADDR=$MONGO_CONTAINER:27017
export MQADDR=$MQ_CONTAINER:5672

# Microservice addresses.
export MESSAGESVCADDR=info-344-messaging:80
export SUMMARYSVCADDR=info-344-summary:80

# Make sure to get the latest image.
docker pull zicodeng/$GATEWAY_CONTAINER

# Remove the old containers first.
if [ "$(docker ps -aq --filter name=$GATEWAY_CONTAINER)" ]; then
    docker rm -f $GATEWAY_CONTAINER
fi

if [ "$(docker ps -aq --filter name=$REDIS_CONTAINER)" ]; then
    docker rm -f $REDIS_CONTAINER
fi

if [ "$(docker ps -aq --filter name=$MONGO_CONTAINER)" ]; then
    docker rm -f $MONGO_CONTAINER
fi

if [ "$(docker ps -aq --filter name=$MQ_CONTAINER)" ]; then
    docker rm -f $MQ_CONTAINER
fi

# Remove dangling images.
if [ "$(docker images -q -f dangling=true)" ]; then
    docker rmi $(docker images -q -f dangling=true)
fi

# Clean up the system.
docker system prune -f

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

# Run Mongo Docker container inside our appnet private network.
docker run \
-d \
-e MONGO_INITDB_DATABASE=$DBNAME \
--name mongo-server \
--network $APP_NETWORK \
--restart unless-stopped \
drstearns/mongo1kusers

# Run RabbitMQ Docker container.
docker run \
-d \
-p 5672:5672 \
--network $APP_NETWORK \
--name $MQ_CONTAINER \
--hostname $MQ_CONTAINER \
rabbitmq

# Run Info 344 API Gateway Docker container inside our appnet private network.
docker run \
-d \
-p 443:443 \
--name $GATEWAY_CONTAINER \
--network $APP_NETWORK \
-v /etc/letsencrypt:/etc/letsencrypt:ro \
-e TLSCERT=$TLSCERT \
-e TLSKEY=$TLSKEY \
-e SESSIONKEY=$SESSIONKEY \
-e ADDR=$ADDR \
-e REDISADDR=$REDISADDR \
-e DBADDR=$DBADDR \
-e MQADDR=$MQADDR \
-e MESSAGESVCADDR=$MESSAGESVCADDR \
-e SUMMARYSVCADDR=$SUMMARYSVCADDR \
--restart unless-stopped \
zicodeng/$GATEWAY_CONTAINER