#!/bin/bash
set -e

docker compose up -d mockmail

sleep 5

docker compose up -d ical-notifier

docker compose exec mockmail cat /var/spool/mail/root

# give the email from root but skip the first 9 lines
diff -w email.eml <(docker compose exec mockmail tail -n +10 /var/spool/mail/root)
