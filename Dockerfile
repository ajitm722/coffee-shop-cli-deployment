# syntax=docker/dockerfile:1

# --- Build stage ---
FROM golang:1.23-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Build a static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/coffee ./cmd/coffee

# --- Runtime stage ---
FROM alpine:3.20
WORKDIR /app
# (Optional) add a non-root user
RUN adduser -D -g '' appuser
COPY --from=builder /out/coffee /app/coffee
USER appuser
EXPOSE 9090
ENTRYPOINT ["/app/coffee","serve"]

