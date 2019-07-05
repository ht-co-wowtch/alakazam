#!/usr/bin/env bash

set -eu -o pipefail

tag="${1:-latest}"

docker build -t alakazams_build:${tag} .