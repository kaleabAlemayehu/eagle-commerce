FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY ./user-ms/go.mod ./user-ms/go.sum ./user-ms/
COPY shared/ ./shared 

WORKDIR /app/user-ms
RUN go mod download

WORKDIR /app
COPY user-ms/ ./user-ms

WORKDIR /app/user-ms
RUN CGO_ENABLED=0 GOOS=linux go build -o user-service ./cmd

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/user-ms/user-service .
EXPOSE 8081

CMD ["./user-service"]
