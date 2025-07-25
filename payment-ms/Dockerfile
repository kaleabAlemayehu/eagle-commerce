FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY ./payment-ms/go.mod ./payment-ms/go.sum ./payment-ms/
COPY shared/ ./shared 

WORKDIR /app/payment-ms
RUN go mod download

WORKDIR /app
COPY payment-ms/ ./payment-ms

WORKDIR /app/payment-ms
RUN CGO_ENABLED=0 GOOS=linux go build -o payment-service ./cmd

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/payment-ms/payment-service .
EXPOSE 8084

CMD ["./payment-service"]
