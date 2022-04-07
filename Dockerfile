FROM golang:alpine AS build

RUN apk --update add git

WORKDIR /app
COPY . /app

RUN go build .

FROM alpine AS run

COPY --from=build /app/ical-relay /app/ical-relay
COPY ./templates/ /app/templates/

WORKDIR /app
VOLUME /app
EXPOSE 80

CMD ./ical-relay