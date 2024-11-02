#!/bin/bash
set -e

docker compose up -d mockmail ical-relay postgres --wait

sleep 5

docker compose up -d ical-notifier

sleep 2

# give the email from root but skip the first 9 lines (since they contain current dates and variable mail transmission data that will change on every run.)
diff -w --strip-trailing-cr  email.eml <(docker compose exec mockmail tail -n +16 /var/spool/mail/root)
