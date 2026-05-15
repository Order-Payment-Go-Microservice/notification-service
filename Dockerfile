FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go mod tidy
RUN go build -o notification-service ./cmd/main.go

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/notification-service .
COPY --from=builder /app/.env .

EXPOSE 9005
EXPOSE 50052

CMD ["./notification-service"]
