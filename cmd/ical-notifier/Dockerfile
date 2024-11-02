FROM golang:1.21-alpine AS build

RUN apk --update add git bash

WORKDIR /app
COPY . /app

RUN go generate ./cmd/ical-notifier/
RUN go build -o ./cmd/ical-notifier/ical-notifier ./cmd/ical-notifier/

FROM alpine AS run

RUN apk --update add ca-certificates curl
COPY --from=build /app/cmd/ical-notifier/ical-notifier /usr/bin/ical-notifier
COPY ./cmd/ical-notifier/misc/ical-notifier.cron /etc/crontabs/root

RUN mkdir -p /etc/ical-relay/notifystore/

WORKDIR /etc/ical-relay
VOLUME /etc/ical-relay/

CMD crond -f -d 8
