FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/skrepka-backend ./cmd/server

FROM alpine:3.22.1

RUN apk add --no-cache curl=8.14.1-r1

WORKDIR /app
RUN mkdir configs

COPY --from=builder /app/skrepka-backend /app/skrepka-backend

COPY --from=builder /app/configs/config.yml /app/configs/config.yml

EXPOSE 4000

CMD ["/app/skrepka-backend"]