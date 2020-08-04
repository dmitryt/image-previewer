# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOLINT=golangci-lint run
GOGET=$(GOCMD) get
BINARY_NAME=image-previewer

all: clean test lint build
test:
	$(GOTEST) -v -race -count 100 ./...
lint:
	$(GOLINT) ./...
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
run:
	$(GOBUILD) -o $(BINARY_NAME) -v ./...
	./$(BINARY_NAME)