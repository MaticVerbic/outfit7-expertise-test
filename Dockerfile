FROM golang:1.14 AS builder
COPY . /app
WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /tmp/api ./cmd/api

FROM alpine:latest
RUN apk update && apk upgrade
WORKDIR /app
COPY --from=builder /tmp/api /app/api
EXPOSE 80
CMD ["/app/api"]
