# STAGE 1: Build
# We use the version matching your go.mod
FROM golang:1.23-alpine AS builder

# Install git (required for fetching some dependencies)
RUN apk add --no-cache git

WORKDIR /app

# Copy go.mod and (implicitly) go.sum to download dependencies
# We do this before copying source code to cache dependencies layer
COPY go.mod ./
# If you have a go.sum, uncomment the next line:
COPY go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the binary named 'main'
# CGO_ENABLED=0 creates a statically linked binary (no external C libs required)
RUN CGO_ENABLED=0 GOOS=linux go build -o main .


# STAGE 2: Run
# We use a tiny Alpine Linux image for the final container
FROM alpine:latest

WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/main .


# Expose the port defined in your app
EXPOSE 8080

# Run the binary
CMD ["./main"]