# Build stage
FROM golang:1.21-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
COPY migrations ./migrations
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/server

# Run stage
FROM alpine:3.19
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /build/migrations ./migrations
EXPOSE 8080
ENTRYPOINT ["./server"]
