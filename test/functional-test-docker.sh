#!/bin/bash
set -ex

function finish {
    set +e
    docker rm -f ${containerDNS}
    docker rm -f ${containerPong}
    docker rm -f ${containerDNSTester}
    rm -rf ".build-artifacts"
}
trap finish EXIT

projectDir="$( cd "$( dirname "${BASH_SOURCE[0]}" )/../" && pwd )"
containerProjectDir="/go/src/github.com/Oppodelldog/docker-dns"
testImage="golang:1.11.5"

echo "building dnsserver before starting tests, since loading dependencies can take a while."

docker run \
 --name "build_dnsserver_${RANDOM}" \
 --rm \
 -i \
 -v ${projectDir}:${containerProjectDir}  \
 -w ${containerProjectDir} \
 ${testImage} \
 bash -c "go get ./... && mkdir -p .build-artifacts && go build -o .build-artifacts/dnsserver dnsserver/cmd/main.go && chmod -R 0777 .build-artifacts"

echo "now go starting the test"

containerDNS=$(
docker run \
 --name dnsserver \
 --rm \
 -d \
 -v ${projectDir}:${containerProjectDir}  \
 -v /var/run/docker.sock:/var/run/docker.sock \
 -w ${containerProjectDir} \
 ${testImage} \
 bash -c "cd dnsserver && ../.build-artifacts/dnsserver"
)

echo "started dnsserver in container ${containerDNS}"

dnsIp=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' ${containerDNS})
echo "dns ip is ${dnsIp}"

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
echo "started test target 'pong' in container ${containerPong}"

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
echo "started functional tests in container ${containerDNSTester}"

docker logs -f ${containerDNSTester}

