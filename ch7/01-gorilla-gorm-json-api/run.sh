#!/usr/bin/env bash
set -e

# trap ctrl-c and call ctrl_c()
trap ctrl_c INT

function ctrl_c() {
    echo "Trapped CTRL-C"
    docker stop postgres
}

# run mongo
docker run --rm --name postgres -e POSTGRES_PASSWORD=postgres -p 5432:5432 -v "$(pwd)/db:/var/lib/postgresql/data" -d postgres

# run go main
go run main.go
