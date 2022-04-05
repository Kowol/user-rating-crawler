#!/usr/bin/env bash

function tearDown {
  docker-compose -f tests/docker-compose.tests.yaml down --remove-orphans -v
}

trap tearDown EXIT

docker-compose -f tests/docker-compose.tests.yaml up -d --build
docker-compose -f tests/docker-compose.tests.yaml run --rm runner go test -p 1 ./... -run "^TestIntegration|TestAcceptance"