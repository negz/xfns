# Compile stage
FROM golang:1.19 AS build-env

WORKDIR /build
COPY . /build

RUN go build -o bin/fn main.go

FROM ubuntu:22.04

COPY --from=build-env /build/bin/fn /usr/local/bin/fn
COPY example/templates /templates

ENTRYPOINT ["/usr/local/bin/fn"]

