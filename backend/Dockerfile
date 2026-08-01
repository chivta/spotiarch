# Dev with hot reload
FROM golang:1.26.5-alpine AS dev
WORKDIR /build
RUN go install github.com/air-verse/air@v1.65.1
COPY go.mod go.sum ./
RUN go mod download
CMD ["air", "-c", ".air.toml"]

# Migrations
FROM golang:1.26.5-alpine AS migrate
WORKDIR /build
RUN go install github.com/pressly/goose/v3/cmd/goose@v3.27.3
COPY migrations ./migrations
CMD ["goose", "up"]

FROM golang:1.26.5-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app ./cmd/spotiarch

# Prod
FROM gcr.io/distroless/static:nonroot
COPY --from=builder /app /app
ENTRYPOINT ["/app"]
