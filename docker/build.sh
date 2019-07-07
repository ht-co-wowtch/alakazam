#!/bin/sh

set -eu -o pipefail

image="${1:-alakazam}"
tag="${2:-latest}"

for name in logic job comet admin;
do
    if [ "$(docker images -q ${image}/${name} 2> /dev/null)" != "" ]; then
        docker rmi $(docker images -q ${image}/${name} 2> /dev/null)
    fi

    cd ${name} && docker build --build-arg image=${image}:${tag} -t ${image}/${name}:${tag} . && cd ../
done