FROM golang:1.23.5-alpine3.20 AS builder
WORKDIR /app
COPY . .

RUN go build -o ./bin/httpserv ./cmd/httpserv

FROM alpine:3
COPY --from=builder /app/bin/httpserv /httpserv
EXPOSE 8080
ENTRYPOINT ["/httpserv"]