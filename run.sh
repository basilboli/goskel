#!/bin/sh

./stop-stack.sh
make build
./run-stack.sh
docker service logs -f goskel_goskel
