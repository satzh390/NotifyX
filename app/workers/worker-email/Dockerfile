# syntax=docker/dockerfile:1.7

FROM golang:1.23 AS builder
WORKDIR /src
COPY . .
WORKDIR /src/app/workers/worker-email
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/worker-email ./cmd/main.go

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /out/worker-email /app/worker-email
COPY --from=builder /src/app/workers/worker-email/config/config.yaml /app/config.yaml
ENV NOTIFYX_WORKER_EMAIL_CONFIG=/app/config.yaml
ENTRYPOINT ["/app/worker-email"]

