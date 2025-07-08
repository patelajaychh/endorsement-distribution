# Copyright 2025 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o endorsement-distribution ./cmd

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/endorsement-distribution .

# Copy config file
COPY --from=builder /app/config/config.yaml ./config/

# Expose port
EXPOSE 8080

# Run the binary
CMD ["./endorsement-distribution"] 