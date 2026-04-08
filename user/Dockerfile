FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -v -o user-service .

FROM alpine:3.18

WORKDIR /app
COPY --from=builder /app/user-service .

ENV PRIVATE_KEY_PATH=/etc/keys/private.pem
EXPOSE 8081

CMD ["/app/user-service"]
