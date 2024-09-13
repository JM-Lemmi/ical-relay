#!/bin/bash
set -e

docker compose up -d

diff -w test.rss <(docker compose exec store cat /etc/ical-relay/rssstore/test.rss)
