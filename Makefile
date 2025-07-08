# Copyright 2025 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

.PHONY: build run test clean

BINARY_NAME=endorsement-distribution
BUILD_DIR=bin

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd

run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME)

test:
	@echo "Running tests..."
	go test ./...

clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)

install-deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

docker-build:
	@echo "Building Docker image..."
	docker build -t $(BINARY_NAME) .

docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 $(BINARY_NAME) 