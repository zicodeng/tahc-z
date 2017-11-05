#!/usr/bin/env bash

# ./build.sh

# Export environment variables.
source devenv.sh

if [ "$(docker ps -aq --filter name=redis-server)" ]; then
    docker rm -f redis-server
fi

# Run Redis Docker container.
docker run \
-d \
--name redis-server \
-p 6379:6379 \
--restart unless-stopped \
redis

if [ "$(docker ps -aq --filter name=mongo-server)" ]; then
    docker rm -f mongo-server
fi

# Run Mongo Docker container.
docker run \
-d \
--name mongo-server \
-p 27017:27017 \
-e MONGO_INITDB_DATABASE=$DBNAME \
--restart unless-stopped \
drstearns/mongo1kusers

# Run API server.
go run main.go