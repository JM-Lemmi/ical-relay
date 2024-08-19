FROM golang:alpine AS build

RUN apk --update add git bash

WORKDIR /app
COPY . /app

RUN go generate ./cmd/ical-relay/
RUN go build -o ./cmd/ical-relay/ical-relay ./cmd/ical-relay/

FROM alpine AS run

RUN apk --update add ca-certificates curl
COPY --from=build /app/cmd/ical-relay/ical-relay /usr/bin/ical-relay

RUN mkdir -p /etc/ical-relay/calstore/
RUN mkdir /etc/ical-relay/notifystore/

COPY cmd/ical-relay/templates/ /opt/ical-relay/templates

WORKDIR /etc/ical-relay
VOLUME /etc/ical-relay/
EXPOSE 80

HEALTHCHECK --interval=60s --start-period=60s --start-interval=5s CMD curl --silent --fail http://localhost:80/health || exit 1

CMD /usr/bin/ical-relay --config /etc/ical-relay/config.yml
