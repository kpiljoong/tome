APP_NAME := tome
CMD_PATH := ./cmd/${APP_NAME}
OUTPUT_DIR := bin
OUTPUT_PATH := $(OUTPUT_DIR)/$(APP_NAME)

GO_FILES := $(shell find . -type f -name '*.go' -not -path "./vendor/*")

.PHONY: all build clean test fmt vet install check fmt-check

all: build

build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(OUTPUT_DIR)
	@go build -o $(OUTPUT_PATH) $(CMD_PATH)

run:
	@echo "Running $(APP_NAME)..."
	@go run $(CMD_PATH)

test:
	@echo "Running tests..."
	@go test -v ./... | tee result.log
	@echo "Tests completed. Check result.log for details."

fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Code formatted."

fmt-check:
	@echo "Checking code format..."
	@go fmt ./... | tee result.log
	@if [ -s result.log ]; then \
		echo "Code is not formatted. Please run 'make fmt' to format the code."; \
		exit 1; \
	else \
		echo "Code is formatted correctly."; \
	fi

vet:
	@echo "Running go vet..."
	@go vet ./...
	@echo "go vet completed."

check: fmt-check vet test

clean:
	@echo "Cleaning up..."
	@rm -rf $(OUTPUT_DIR)
	@echo "Cleaned up."

install:
	@echo "Installing $(APP_NAME)..."
	@go install $(CMD_PATH)
	@echo "$(APP_NAME) installed."
