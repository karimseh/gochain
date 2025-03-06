BIN_NAME := gochain
BIN_DIR := bin

.PHONY: build clean

build:
	@echo "Building binary..."
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/$(BIN_NAME) cmd/gochain/main.go
	@echo "Binary created: $(BIN_DIR)/$(BIN_NAME)"

clean:
	@echo "Cleaning binaries..."
	@rm -rf $(BIN_DIR)

test:
	@echo "Running tests..."
	@go test -v ./...

run: build
	@echo "Starting application..."
	@./$(BIN_DIR)/$(BIN_NAME)