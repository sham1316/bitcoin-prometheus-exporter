FROM golang:1.15.5 AS builder

WORKDIR /app/

COPY go.* /app/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go mod download

COPY . /app/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bitcoind_exporter .

FROM alpine:3.12.0
WORKDIR /app
COPY --from=builder /app/bitcoind_exporter .

ENTRYPOINT ["/app/bitcoind_exporter"]




