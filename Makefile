# Define the name of the binary
BINARY_NAME=btc-node

# The directory where the main.go file is located
SOURCE_DIR=./cmd/btc-node/

# The output directory for the binary
OUTPUT_DIR=./bin

# The Go build command
BUILD_CMD=go build -o $(OUTPUT_DIR)/$(BINARY_NAME) $(SOURCE_DIR)

# The Go run command
#RUN_CMD=$(OUTPUT_DIR)/btc-node -config=$(OUTPUT_DIR)

.PHONY: all build run clean

# The default task
all: build

# Task to build the program
build:
	@echo "Building the Go program..."
	@mkdir -p $(OUTPUT_DIR)
	$(BUILD_CMD)

# Task to run the program
run:
	@echo "Running the Go program..."
	$(RUN_CMD)

# Task to clean up the output directory
clean:
	@echo "Cleaning up..."
	@rm -rf $(OUTPUT_DIR)

# Task to install dependencies (optional)
deps:
	@echo "Installing dependencies..."
	@go mod tidy
