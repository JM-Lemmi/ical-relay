FROM golang:alpine AS build

RUN apk --update add git

WORKDIR /etc/ical-relay
COPY . /etc/ical-relay

RUN go build .

FROM alpine AS run

COPY --from=build /app/ical-relay /usr/bin/ical-relay
RUN mkdir /etc/ical-relay/calstore/
RUN mkdir /etc/ical-relay/notifystore/

WORKDIR /etc/ical-relay
VOLUME /etc/ical-relay/
EXPOSE 80

CMD /usr/bin/ical-relay --config /etc/ical-relay/config.yml
