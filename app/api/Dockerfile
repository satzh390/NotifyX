# syntax=docker/dockerfile:1.7

FROM golang:1.23 AS builder
WORKDIR /src
COPY . .
WORKDIR /src/app/api
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/notifyx-api ./cmd/main.go

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /out/notifyx-api /app/notifyx-api
COPY --from=builder /src/app/api/config/config.yaml /app/config.yaml
ENV NOTIFYX_API_CONFIG=/app/config.yaml
EXPOSE 8080
ENTRYPOINT ["/app/notifyx-api"]

