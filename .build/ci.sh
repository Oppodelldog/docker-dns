#!/usr/bin/env bash

set -ex

projectDir="$( cd "$( dirname "${BASH_SOURCE[0]}" )/../" && pwd )"

${projectDir}/test/functional-test-docker.sh

