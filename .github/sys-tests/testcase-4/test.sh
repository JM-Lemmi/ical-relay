#!/bin/bash
set -e

docker compose up -d mockmail

sleep 5

docker compose up -d ical-notifier

# give the email from root but skip the first 9 lines (since they contain current dates and variable mail transmission data that will change on every run.)
diff -w --strip-trailing-cr  email.eml <(docker compose exec mockmail tail -n +15 /var/spool/mail/root)
