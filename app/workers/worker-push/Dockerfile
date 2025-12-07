# syntax=docker/dockerfile:1.7

FROM golang:1.23 AS builder
WORKDIR /src
COPY . .
WORKDIR /src/app/workers/worker-push
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/worker-push ./cmd/main.go

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /out/worker-push /app/worker-push
COPY --from=builder /src/app/workers/worker-push/config/config.yaml /app/config.yaml
ENV NOTIFYX_WORKER_PUSH_CONFIG=/app/config.yaml
ENTRYPOINT ["/app/worker-push"]

