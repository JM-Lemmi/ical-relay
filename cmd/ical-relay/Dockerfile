FROM golang:alpine AS build

RUN apk --update add git

WORKDIR /app
COPY . /app

RUN if [[ $(git tag -l --contains HEAD) ]]; then echo -n $(git tag -l --contains HEAD) > ./cmd/ical-relay/VERSION; else echo -n $(git rev-parse --short HEAD) > ./cmd/ical-relay/VERSION; fi

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
