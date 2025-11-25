# syntax=docker/dockerfile:1.7

FROM golang:1.23 AS builder
WORKDIR /src
COPY . .
WORKDIR /src/app/processor
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/notifyx-processor ./cmd/main.go

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /out/notifyx-processor /app/notifyx-processor
COPY --from=builder /src/app/processor/config/config.yaml /app/config.yaml
ENV NOTIFYX_PROCESSOR_CONFIG=/app/config.yaml
ENTRYPOINT ["/app/notifyx-processor"]

