# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY services/leaderboard-service/go.mod services/leaderboard-service/go.mod
COPY services/leaderboard-service/go.sum services/leaderboard-service/go.sum

COPY common /app/common

WORKDIR /app/services/leaderboard-service

RUN go mod download

COPY services/leaderboard-service/ .

RUN go build -o leaderboard-service ./cmd

# Run stage
FROM alpine:3.19
WORKDIR /app

COPY --from=builder /app/services/leaderboard-service .

EXPOSE 9092
CMD ["/app/leaderboard-service"]