# -- Stage 1: build ---
FROM golang:1.24.5-alpine3.22 AS builder

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR /app


COPY go.mod go.sum ./

RUN go mod download


COPY ../.. .

# Build the binary
RUN go build -o hcaas ./cmd/url

# --- Stage 2: minimal runtime ---
FROM gcr.io/distroless/static-debian11:nonroot


WORKDIR /

COPY --from=builder /app/hcaas .

EXPOSE 8080

ENTRYPOINT ["/hcaas"]
