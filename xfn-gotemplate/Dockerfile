# Compile stage
FROM golang:1.19 AS build-env

WORKDIR /build
COPY . /build

RUN go build -o bin/fn main.go

FROM gcr.io/distroless/static@sha256:11364b4198394878b7689ad61c5ea2aae2cd2ed127c09fc7b68ca8ed63219030

COPY --from=build-env /build/bin/fn /usr/local/bin/fn
COPY example/templates /templates

ENTRYPOINT ["/usr/local/bin/fn"]

