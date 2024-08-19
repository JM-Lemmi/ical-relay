#!/bin/bash
set -e

docker compose up --build -d --wait

curl -X POST http://localhost:8080/api/profiles/test -H "Authorization: supersecret" --data "{\"sources\":[\"base64://$(base64 -w0 ../../testdata/basic.ics)\"],\"public\": false,\"immutable-past\": true}"

diff -w ../../testdata/basic.ics <(curl -s http://localhost:8080/profiles/test)
