#!/bin/bash
set -e

docker compose up -d --wait

diff -w ../../testdata/basic.ics <(curl -s http://localhost:8080/profiles/test)
