# Simple Go HTTP server example for custom components
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files if they exist
COPY go.mod go.sum* ./
RUN if [ -f go.mod ]; then go mod download; fi

# Copy source code
COPY main.go .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o app main.go

# Final stage - minimal runtime image
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/app .

# Expose port 8080
EXPOSE 8080

# Run the application
CMD ["./app"]
