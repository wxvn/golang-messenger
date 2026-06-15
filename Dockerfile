FROM golang:1.26.3 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o messenger ./cmd/messenger

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/messenger .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/docs ./docs

RUN mkdir -p logs

EXPOSE 8080

CMD ["./messenger"]
