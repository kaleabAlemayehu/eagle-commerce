FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY product-ms/go.mod product-ms/go.sum ./product-ms/
COPY shared/ ./shared 

WORKDIR /app/product-ms
RUN go mod download

WORKDIR /app
COPY product-ms/ ./product-ms

WORKDIR /app/product-ms
RUN CGO_ENABLED=0 GOOS=linux go build -o product-service ./cmd

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/product-ms/product-service .
EXPOSE 8082

CMD ["./product-service"]
