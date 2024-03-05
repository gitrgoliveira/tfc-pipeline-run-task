.PHONY: all clean

# The name of the Go executable
BINARY_NAME := tfc-pipeline-run-task

all: clean build run

build: clean
	@echo "Compiling..."
	@go build -buildvcs=false -o $(BINARY_NAME) .

run: clean build
	@echo "Running..."
	@./$(BINARY_NAME) serve

clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)

buildDocker:
	@docker build -t $(BINARY_NAME) .
	
runDocker: buildDocker
	@docker run -p 80:80 $(BINARY_NAME)
