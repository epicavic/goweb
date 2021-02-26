#!/usr/bin/env bash

# trap ctrl-c and call ctrl_c()
trap ctrl_c INT

function ctrl_c() {
    echo "Trapped CTRL-C"
    docker stop rabbitmq
    docker stop redis
    exit
}

# run containers
docker run --rm --name rabbitmq -p 5672:5672 -d rabbitmq
docker run --rm --name redis -p 6379:6379 -d redis

# run go main
go run main.go

# stop container
docker stop rabbitmq
docker stop redis
