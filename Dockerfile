# Build stage
FROM golang:1.22-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# Final stage
FROM alpine:latest

# Install ca-certificates and tzdata for timezone support
RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -S appgroup && \
    adduser -S appuser -G appgroup && \
    mkdir -p /app/web /app/migrations && \
    chown -R appuser:appgroup /app

# Set timezone to Eastern Standard Time
ENV TZ=America/New_York

# Set working directory
WORKDIR /app

# Copy binary from builder and set permissions
COPY --from=builder --chown=appuser:appgroup /app/server .

# Copy web files and migrations with proper permissions
COPY --chown=appuser:appgroup web/ web/
COPY --chown=appuser:appgroup migrations/ migrations/

# Create directories for secrets and migrations, ensure proper permissions
RUN mkdir -p /app/secrets /app/migrations && \
    chown -R appuser:appgroup /app/secrets /app/migrations && \
    chmod 700 /app/secrets && \
    chmod 755 /app/migrations

# Copy .env file for secrets with restricted permissions
COPY --chown=appuser:appgroup .env /app/secrets/.env
RUN chmod 600 /app/secrets/.env

# Expose port
EXPOSE 8080

# Switch to non-root user
USER appuser

# Set resource limits
ENV GOMAXPROCS=2

# Set security options
SECURITY_OPTS="--security-opt=no-new-privileges --read-only --cap-drop=ALL"

# Run the application
CMD ["./server"]