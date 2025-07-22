FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/skrepka-backend ./cmd/server

FROM alpine:3.22.1

RUN apk add --no-cache curl=8.15.0-r0

COPY --from=builder /app/skrepka-backend /app/skrepka-backend

EXPOSE 4000

CMD ["/app/skrepka-backend"]