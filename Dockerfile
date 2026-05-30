# syntax=docker/dockerfile:1

FROM golang:1.22-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/alert-service ./main.go

FROM alpine:3.20

RUN addgroup -S app && adduser -S app -G app

WORKDIR /app

ENV GIN_MODE=release
ENV PORT=8080

COPY --from=builder /out/alert-service /app/alert-service

USER app

EXPOSE 8080

ENTRYPOINT ["/app/alert-service"]
