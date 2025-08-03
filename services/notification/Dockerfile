# --- Stage 1: Build the application ---
FROM golang:1.24.5-alpine3.22 AS builder

# Set build environment variables for a static binary.
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

# Set the working directory inside the container.
WORKDIR /app

# Copy go.mod and go.sum to cache dependencies.
COPY go.mod ./
COPY go.sum ./

# Download all dependencies.
RUN go mod download

# Copy the rest of the application source code.
COPY . .

# Build the application binary. The output is named `notification-service`
# and is placed in the working directory (/app).
RUN go build -o notification-service ./cmd/notification

# --- Stage 2: Create the final, minimal image ---
# Use a distroless base image for a small, secure, and production-ready image.
FROM gcr.io/distroless/static-debian11:nonroot

# Set the working directory to root.
WORKDIR /

# Copy the compiled binary from the builder stage.
COPY --from=builder /app/notification-service /notification-service

# Expose the port that the health check server will run on (8082).
EXPOSE 8082

# Set the entrypoint to run the binary.
ENTRYPOINT ["/notification-service"]
