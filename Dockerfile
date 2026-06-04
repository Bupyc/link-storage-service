FROM golang:1.25.3-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o link-storage-service ./cmd/app


FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/link-storage-service .

EXPOSE 8080

CMD ["./link-storage-service"]