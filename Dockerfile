# Use the official Golang image as the build environment
FROM golang:1.22.1 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the application code
COPY . .

# Build the Go application
RUN go build -o vault-signer .

# Use a minimal base image for the final executable
FROM gcr.io/distroless/base-debian11

# Set the working directory in the final container
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/vault-signer /app/vault-signer

# Expose port if needed (optional)
EXPOSE 8200

# Run the application
ENTRYPOINT ["/app/vault-signer"]