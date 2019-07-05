#!/usr/bin/env bash

set -eu -o pipefail

tag="${1:-latest}"

for name in logic job comet admin;
do
    if [ "$(docker images -q alakazam_${name} 2> /dev/null)" != "" ]; then
        docker rmi $(docker images -q alakazam_${name} 2> /dev/null)
    fi

    cd ${name} && docker build -t alakazam_${name}:${tag} . && cd ../
done