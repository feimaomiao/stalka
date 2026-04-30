# syntax=docker/dockerfile:1.7

# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-w -s" -o /out/stalka .

# Runtime stage
FROM alpine:3.21

RUN apk --no-cache add ca-certificates tzdata \
    && addgroup -S app && adduser -S -G app -u 10001 app

WORKDIR /app

COPY --from=builder /out/stalka /app/stalka

USER app

ENTRYPOINT ["/app/stalka"]
