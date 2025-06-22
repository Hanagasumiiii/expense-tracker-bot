FROM golang:1.24-alpine AS builder
WORKDIR /app

RUN apk add --no-cache git make

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o bot ./cmd/bot

FROM alpine:3.20
WORKDIR /app

COPY --from=builder /app/bot ./bot

RUN apk add --no-cache ca-certificates

RUN adduser -D -g '' botuser
USER botuser

ENTRYPOINT ["./bot"]