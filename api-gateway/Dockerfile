FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY api-gateway/go.mod api-gateway/go.sum ./api-gateway/
COPY shared/go.mod shared/go.sum ./shared/

WORKDIR /app/api-gateway
RUN go mod download

WORKDIR /app
COPY api-gateway/ ./api-gateway/
COPY shared/ ./shared/

WORKDIR /app/api-gateway
RUN CGO_ENABLED=0 GOOS=linux go build -o api-gateway ./cmd

FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/api-gateway/api-gateway .

EXPOSE 8080

CMD ["./api-gateway"]
