# syntax=docker/dockerfile:1.7

FROM golang:1.23 AS builder
WORKDIR /src
COPY . .
WORKDIR /src/app/workers/worker-webhook
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/worker-webhook ./cmd/main.go

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /out/worker-webhook /app/worker-webhook
COPY --from=builder /src/app/workers/worker-webhook/config/config.yaml /app/config.yaml
ENV NOTIFYX_WORKER_WEBHOOK_CONFIG=/app/config.yaml
ENTRYPOINT ["/app/worker-webhook"]

