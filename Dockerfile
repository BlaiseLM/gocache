# Official golang iamge (version 1.23) running on alpine distro
FROM golang:1.23-alpine AS builder

# /build as working directory
WORKDIR /build 

# Copy dependencies & checksums of dependencies
COPY go.mod go.sum ./

# Download dependencies in working directory
RUN go mod download

# Copy all files into working directory
COPY . .

# Build gocache instance
RUN go build -o gocache

# Only gocache executable
FROM alpine:latest
WORKDIR /app
COPY --from=builder /build/gocache .
COPY --from=builder /build/.env .

# TCP server
EXPOSE 8080

# Prometheus metrics
EXPOSE 8081

CMD ["./gocache"]