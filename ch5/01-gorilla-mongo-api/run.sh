#!/usr/bin/env bash

# trap ctrl-c and call ctrl_c()
trap ctrl_c INT

function ctrl_c() {
    echo "Trapped CTRL-C"
    docker stop mongo
    exit
}

# run mongo
docker run --rm --name mongo -p 27017:27017 -v "$(pwd)/db:/data/db" -d mongo

# run go main
go run main.go

# stop container
docker stop mongo
