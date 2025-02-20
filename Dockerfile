# Use the official Golang Alpine base image
FROM golang:1.21-alpine AS builder

# Set the working directory
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy Go module files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the Go application
RUN go build -o webfinger

# Create a minimal runtime image
FROM alpine:latest

# Set working directory
WORKDIR /app

# Install required packages (optional: ca-certificates for HTTPS)
RUN apk add --no-cache ca-certificates

# Copy the binary from the builder stage
COPY --from=builder /app/webfinger /app/webfinger

# Copy the config file (ensure it's in the correct location)
COPY config.yaml /app/config.yaml

# Expose the WebFinger service port
EXPOSE 8000

# Run the WebFinger server
CMD ["/app/webfinger"]

