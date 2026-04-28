FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/server/main.go

FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add tzdata

COPY --from=builder /app/server .
COPY migrations /app/migrations

CMD ["./server"]
