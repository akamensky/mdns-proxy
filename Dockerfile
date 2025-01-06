FROM golang:1.23-alpine3.20 as builder

ADD . /code
# Run tests
RUN go env && cd /code && go test -buildvcs=false ./...
# Compile the binary
RUN go env && cd /code && go build -buildvcs=false -o /mdns-proxy .

FROM alpine:3.20

LABEL org.opencontainers.image.source=https://github.com/akamensky/mdns-proxy

COPY --from=builder /mdns-proxy /bin/mdns-proxy

ENTRYPOINT ["mdns-proxy"]
