#!/bin/sh

. ./.env && docker stack deploy --with-registry-auth --compose-file docker-compose-prod.yml goskel