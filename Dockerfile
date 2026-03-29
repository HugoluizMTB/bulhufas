FROM golang:1.24-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build \
    -ldflags="-w -s -linkmode external -extldflags '-static'" \
    -trimpath \
    -o /app/bulhufas ./cmd/server

FROM alpine:3.20

RUN adduser -D -u 1000 bulhufas
COPY --from=builder /app/bulhufas /usr/local/bin/bulhufas

EXPOSE 8420
VOLUME /data
ENV DATA_DIR=/data

USER bulhufas
ENTRYPOINT ["bulhufas"]
