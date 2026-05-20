FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY proto-generation/ /app/proto-generation/

WORKDIR /app/notification-service
COPY notification-service/go.mod notification-service/go.sum ./
RUN go mod download

COPY notification-service/ .
RUN go build -o notification-service ./cmd/main.go

FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /root/
COPY --from=builder /app/notification-service/notification-service .
EXPOSE 9005
EXPOSE 50055
CMD ["./notification-service"]
