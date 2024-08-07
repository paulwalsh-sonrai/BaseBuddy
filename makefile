# Define variables
IMAGE_NAME=basebuddy
CONTAINER_NAME=basebuddy-container
PORT=8080

# Default target
.PHONY: all
all: build

# Build the Docker image
.PHONY: build
build:
	@echo "Building Docker image..."
	docker build -t $(IMAGE_NAME) .

# Run the Docker container
.PHONY: run
run:
	@echo "Running Docker container..."
	docker run -d -p $(PORT):8080 --name $(CONTAINER_NAME) $(IMAGE_NAME)

# Stop the Docker container
.PHONY: stop
stop:
	@echo "Stopping Docker container..."
	docker stop $(CONTAINER_NAME)

# Remove the Docker container
.PHONY: clean
clean: stop
	@echo "Removing Docker container..."
	docker rm $(CONTAINER_NAME)

# Remove the Docker image
.PHONY: rmi
rmi: clean
	@echo "Removing Docker image..."
	docker rmi $(IMAGE_NAME)

# Rebuild and run the Docker container
.PHONY: rebuild
rebuild: rmi build run
	@echo "Rebuilding and running Docker container..."

