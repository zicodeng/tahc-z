#!/usr/bin/env bash

# Server address.
export ADDR=localhost:443

export TLSCERT="/c/Users/Zico Deng/Desktop/go/src/github.com/info344-a17/challenges-zicodeng/tls/fullchain.pem"
export TLSKEY="/c/Users/Zico Deng/Desktop/go/src/github.com/info344-a17/challenges-zicodeng/tls/privkey.pem"

export SESSIONKEY="secret signing key"

export REDISADDR=192.168.99.100:6379
export DBADDR=192.168.99.100:27017
export DBNAME="info_344"

export MESSAGESVCADDR=localhost:4000
export SUMMARYSVCADDR=localhost:5000
