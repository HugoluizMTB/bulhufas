FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -trimpath \
    -o /app/bulhufas ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /app/bulhufas /bulhufas

EXPOSE 8420

USER nonroot:nonroot
ENTRYPOINT ["/bulhufas"]
