# Build stage
FROM golang:1.25-alpine AS builder
WORKDIR /app

COPY ../../go.mod ../../go.sum ./

COPY ../../ ./

WORKDIR /app/services/leaderboard-service

RUN go mod tidy
RUN go build -o leaderboard-service

# Run stage
FROM alpine:3.19
WORKDIR /app

COPY --from=builder /app/services/leaderboard-service/leaderboard-service .

EXPOSE 8010
CMD ["./leaderboard-service"]