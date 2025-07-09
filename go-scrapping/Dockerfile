# Stage 1: Build the Go binary
FROM golang:1.23-alpine AS builder

# Set environment variables for Go
ENV CGO_ENABLED=0 \
  GOOS=linux \
  GOARCH=amd64

# Set working directory
WORKDIR /app

# Copy Go mod files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project and build it
COPY . .
RUN go build -o scraper .

# Stage 2: Create a minimal runtime container
FROM alpine:latest

# Set working directory in final container
WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/scraper .

# Command to run the scraper
CMD ["./scraper"]

