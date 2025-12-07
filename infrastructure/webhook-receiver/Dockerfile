FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY main.go .
RUN go mod init webhook-receiver && \
    go get github.com/gorilla/mux && \
    go build -o webhook-receiver .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/webhook-receiver .
EXPOSE 8888
CMD ["./webhook-receiver"]

