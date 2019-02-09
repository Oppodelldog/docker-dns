#!/usr/bin/env bash

set -ex

projectDir="$( cd "$( dirname "${BASH_SOURCE[0]}" )/../" && pwd )"

cd ${projectDir}/test || exit 1

./functional-test-docker.sh

