# Используем официальный образ Go
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go install github.com/swaggo/swag/cmd/swag@latest

RUN swag init -g cmd/main.go

RUN go build -o app ./cmd/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/app .
COPY --from=builder /app/docs ./docs
COPY --from=builder /app/.env .env

# Устанавливаем инструменты для анализа pprof
RUN apk add --no-cache go bash curl

EXPOSE 8080

CMD ["./app"]