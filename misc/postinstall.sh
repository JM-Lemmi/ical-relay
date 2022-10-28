#!/bin/sh
if ! getent group ical-relay > /dev/null 2>&1 ; then
    addgroup --system ical-relay --quiet
fi
if ! id ical-relay > /dev/null 2>&1 ; then
    adduser --system --no-create-home \
        --ingroup ical-relay --disabled-password --shell /bin/false \
        ical-relay
fi

systemctl enable ical-relay.service
