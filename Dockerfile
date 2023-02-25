FROM golang:alpine AS build

RUN apk --update add git

WORKDIR /app
COPY . /app

RUN go build .

FROM alpine AS run

COPY --from=build /app/ical-relay /usr/bin/ical-relay
RUN mkdir -p /etc/ical-relay/calstore/
RUN mkdir /etc/ical-relay/notifystore/

COPY templates/ /opt/ical-relay/templates

WORKDIR /etc/ical-relay
VOLUME /etc/ical-relay/
EXPOSE 80

CMD /usr/bin/ical-relay --config /etc/ical-relay/config.yml
