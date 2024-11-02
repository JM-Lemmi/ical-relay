#!/bin/bash
set -e

docker compose up -d mockhook

sleep 5

docker compose up -d ical-notifier

diff -w body.json <(docker compose exec mockhook cat /var/www/html/body.json)
