FROM golang:1.20-alpine as builder

WORKDIR /app

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY cmd/ cmd/
COPY internal/ internal/
COPY pkg/ pkg/

RUN CGO_ENABLED=0 GOOS=linux go build -o /function cmd/main.go

###
FROM alpine:latest

COPY crossplane.yaml /crossplane.yaml
COPY --from=builder /function /function

ENTRYPOINT ["/function"]
