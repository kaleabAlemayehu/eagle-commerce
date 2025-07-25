FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY ./order-ms/go.mod ./order-ms/go.sum ./order-ms/
COPY shared/ ./shared 

WORKDIR /app/order-ms
RUN go mod download

WORKDIR /app
COPY order-ms/ ./order-ms 

WORKDIR /app/order-ms
RUN CGO_ENABLED=0 GOOS=linux go build -o order-service ./cmd

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/order-ms/order-service .
EXPOSE 8083

CMD ["./order-service"]
