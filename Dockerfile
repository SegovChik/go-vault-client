# Use the official Golang image as the build environment
FROM go:1.22.8-bullseye AS builder
# Set the working directory inside the container
WORKDIR /app
# Copy go.mod and go.sum files and download dependencies
COPY go.mod go.sum ./
RUN go mod download
# Copy the application code
COPY . .
# Build the Go application with static linking
RUN CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' -o vault-signer .

FROM go:1.22.8-bullseye
# Set the working directory in the final container
WORKDIR /app
# Copy the binary from the builder stage
COPY --from=builder /app/vault-signer /app/vault-signer
# Run the application
ENTRYPOINT ["/app/vault-signer"]
