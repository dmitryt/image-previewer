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
	CACHE_DIR=.cache-test $(GOTEST) -v `go list ./... | grep -v integration-tests`
test-race:
	CACHE_DIR=.cache-test $(GOTEST) -v -race -count 100 `go list ./... | grep -v integration-tests`
test-integration:
	docker-compose -f integration-tests/docker-compose.yml up -d --build
	$(GOTEST) -v ./integration-tests/...
	docker-compose -f integration-tests/docker-compose.yml down
lint:
	$(GOLINT) ./...
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
run:
	$(GOBUILD) -o $(BINARY_NAME)
	./$(BINARY_NAME)