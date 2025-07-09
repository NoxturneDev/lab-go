# Use official Golang image
FROM golang:1.22-alpine

# Set working directory inside the container
WORKDIR /app

# Copy the Go code
COPY . .

# Build the Go app
RUN go build -o server main.go

# Expose port 8080
EXPOSE 8080

# Run the built binary
CMD ["./server"]

