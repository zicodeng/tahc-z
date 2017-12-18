#!/usr/bin/env bash

# Export environment variables.
export ADDR=localhost:443

export TLSCERT="/c/Users/Zico Deng/Desktop/go/src/github.com/info344-a17/challenges-zicodeng/tls/fullchain.pem"
export TLSKEY="/c/Users/Zico Deng/Desktop/go/src/github.com/info344-a17/challenges-zicodeng/tls/privkey.pem"

export SESSIONKEY="secret signing key"

export REDISADDR=192.168.99.100:6379
export MQADDR=192.168.99.100:5672
export DBADDR=192.168.99.100:27017
export DBNAME="info_344"

export REDIS_CONTAINER=redis-server
export MONGO_CONTAINER=mongo-server
export MQ_CONTAINER=rabbitmq-server

export MESSAGESVCADDR=localhost:4000
export SUMMARYSVCADDR=localhost:5000

if [ "$(docker ps -aq --filter name=$REDIS_CONTAINER)" ]; then
    docker rm -f $REDIS_CONTAINER
fi

# Run Redis Docker container.
docker run \
-d \
--name $REDIS_CONTAINER \
-p 6379:6379 \
redis

if [ "$(docker ps -aq --filter name=$MONGO_CONTAINER)" ]; then
    docker rm -f $MONGO_CONTAINER
fi

# Run Mongo Docker container.
docker run \
-d \
--name $MONGO_CONTAINER \
-p 27017:27017 \
-e MONGO_INITDB_DATABASE=$DBNAME \
drstearns/mongo1kusers

if [ "$(docker ps -aq --filter name=$MQ_CONTAINER)" ]; then
    docker rm -f $MQ_CONTAINER
fi

# Run RabbitMQ Docker container.
docker run \
-d \
-p 5672:5672 \
--name $MQ_CONTAINER \
--hostname $MQ_CONTAINER \
rabbitmq

# Run API gateway.
go run main.go