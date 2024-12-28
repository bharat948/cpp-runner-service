# Stage 1: Build the Go application
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc g++ musl-dev

# Set working directory
WORKDIR /app

# Copy only the dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-w -s" -o main .

# Stage 2: Create the final minimal image
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache g++

# Create a non-root user
RUN adduser -D appuser

# Copy the binary from builder
COPY --from=builder /app/main /app/main

# Set ownership
RUN chown appuser:appuser /app/main

# Switch to non-root user
USER appuser

# Expose service port
EXPOSE 8080

# Run the binary
CMD ["/app/main"]