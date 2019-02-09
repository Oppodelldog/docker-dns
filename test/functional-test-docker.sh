#!/bin/bash
set -ex

projectDir="$( cd "$( dirname "${BASH_SOURCE[0]}" )/../" && pwd )"
containerProjectDir="/go/src/github.com/Oppodelldog/docker-dns"
testImage="golang:1.11.0"

containerDNS=$(
docker run \
 --name dnsserver \
 --rm \
 -d \
 -v ${projectDir}:${containerProjectDir}  \
 -v /var/run/docker.sock:/var/run/docker.sock \
 -w ${containerProjectDir}/dnsserver \
 ${testImage} \
 bash -c "go run cmd/main.go"
)

dnsIp=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' ${containerDNS})

containerPong=$(
docker run \
 --name pong \
 --rm \
 -d \
 -v ${projectDir}:${containerProjectDir} \
 -w ${containerProjectDir}  \
 ${testImage} \
 bash -c "go run test/pong/main.go"
 )

containerDNSTester=$(
docker run \
 --name dnstester \
 --dns=${dnsIp} \
 --rm \
 -d \
 -v ${projectDir}:${containerProjectDir}  \
 -w ${containerProjectDir}  \
 ${testImage} \
 bash -c "go run test/dnslookup/main.go"
)

docker logs -f ${containerDNSTester}

