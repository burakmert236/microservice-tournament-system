# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY services/user-service/go.mod services/user-service/go.mod
COPY services/user-service/go.sum services/user-service/go.sum

COPY common /app/common

WORKDIR /app/services/user-service

RUN go mod download

COPY services/user-service/ .

RUN go build -o user-service ./cmd

# Run stage
FROM alpine:3.19
WORKDIR /app

COPY --from=builder /app/services/user-service .

EXPOSE 9090
CMD ["/app/user-service"]