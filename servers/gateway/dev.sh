#!/usr/bin/env bash

TLSCERT="$(pwd)/tls/fullchain.pem" \
TLSKEY="$(pwd)/tls/privkey.pem" \
go run main.go