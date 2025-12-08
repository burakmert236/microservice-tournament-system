# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY services/tournament-service/go.mod services/tournament-service/go.mod
COPY services/tournament-service/go.sum services/tournament-service/go.sum

COPY common /app/common

WORKDIR /app/services/tournament-service

RUN go mod download

COPY services/tournament-service/ .

RUN go build -o tournament-service ./cmd

# Run stage
FROM alpine:3.19
WORKDIR /app

COPY --from=builder /app/services/tournament-service .

EXPOSE 9091
CMD ["/app/tournament-service"]