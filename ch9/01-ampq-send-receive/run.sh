#!/usr/bin/env bash
set -e

# trap ctrl-c and call ctrl_c()
trap ctrl_c INT

function ctrl_c() {
    echo "Trapped CTRL-C"
    docker stop rabbitmq
}

# run container
docker run --rm --name rabbitmq -p 5672:5672 -d rabbitmq

# run go main
go run main.go

# stop container
docker stop rabbitmq
