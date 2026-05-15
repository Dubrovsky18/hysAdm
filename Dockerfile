# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o panel ./cmd/panel

FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/panel .
COPY --from=builder /app/config.yaml .
COPY web ./web

EXPOSE 8080

CMD ["./panel"]
