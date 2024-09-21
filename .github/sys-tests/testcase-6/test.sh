#!/bin/bash
set -e

docker compose up -d

diff -w -I '<pubDate>' test.rss <(docker compose exec store cat /etc/ical-relay/rssstore/test.rss)
