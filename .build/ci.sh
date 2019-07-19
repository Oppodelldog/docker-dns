#!/usr/bin/env bash

docker run --name="docker-dns-go" -v ${PWD}:/app -w /app golang:latest go build -o tests .test/main.go

./tests