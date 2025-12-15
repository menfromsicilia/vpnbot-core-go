# Build stage
FROM golang:1.22-alpine AS builder

# Use direct download instead of proxy to avoid 403 errors
ENV GOPROXY=direct

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git gcc musl-dev sqlite-dev

# Copy everything
COPY . .

# Download dependencies and build
RUN go mod tidy
RUN go mod download
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o vpnbot-core ./cmd/server/main.go

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Copy binary and migrations
COPY --from=builder /app/vpnbot-core .
COPY --from=builder /app/migrations ./migrations

# Create volume for database
VOLUME /app/data

# Expose port
EXPOSE 8080

# Run
CMD ["./vpnbot-core"]

