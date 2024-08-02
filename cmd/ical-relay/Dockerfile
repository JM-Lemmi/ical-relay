FROM golang:alpine AS build

RUN apk --update add git bash

WORKDIR /app
COPY . /app

RUN go generate ./cmd/ical-relay/
RUN go build -o ./cmd/ical-relay/ical-relay ./cmd/ical-relay/

FROM alpine AS run

COPY --from=build /app/cmd/ical-relay/ical-relay /usr/bin/ical-relay
RUN mkdir -p /etc/ical-relay/calstore/
RUN mkdir /etc/ical-relay/notifystore/

COPY cmd/ical-relay/templates/ /opt/ical-relay/templates

WORKDIR /etc/ical-relay
VOLUME /etc/ical-relay/
EXPOSE 80

CMD /usr/bin/ical-relay --config /etc/ical-relay/config.yml
