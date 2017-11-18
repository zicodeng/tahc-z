#!/usr/bin/env bash

# source this script first.

# Docker image name we will build.
export INFO_344_MYSQL_IMAGE=zicodeng/info-344-mysql-server

# Docker container name.
export CONTAINER_NAME=info-344-mysql-server

# Database name in which our schema will be created.
export MYSQL_DATABASE=info_344

# Random MySQL root password.
export MYSQL_ROOT_PASSWORD=$(openssl rand -base64 18)

echo "mysql root password:" $MYSQL_ROOT_PASSWORD

# For Windows Home user, use Linux VM IP address.
export MYSQLADDR=192.168.99.100:3306
