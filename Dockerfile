FROM golang:alpine AS build

RUN apk --update add git

WORKDIR /app
COPY . /app

RUN go build .

FROM alpine AS run

COPY --from=build /app/ical-relay /app/ical-relay
COPY ./templates/ /app/templates/
RUN mkdir /app/calstore/
RUN mkdir /app/notifystore/

WORKDIR /app
VOLUME /app/calstore/
EXPOSE 80

CMD ./ical-relay