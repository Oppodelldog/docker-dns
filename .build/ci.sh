#!/usr/bin/env bash

APP_DIR="$(cd -P "$(dirname "${filename}")/..";pwd)"

docker run --rm --name="docker-dns-go" -v ${APP_DIR}:/app -w /app golang:latest bash -c "pwd && ls -al && go build -o tests .test/main.go"

./tests