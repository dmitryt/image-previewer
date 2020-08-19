# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOLINT=golangci-lint run
GOGET=$(GOCMD) get
BINARY_NAME=previewer
BINARY_DIR=tmp-bin

all: clean test lint build
test:
	CACHE_DIR=.cache-test $(GOTEST) -v `go list ./... | grep -v integration-tests`
test-race:
	CACHE_DIR=.cache-test $(GOTEST) -v -race -count 20 `go list ./... | grep -v integration-tests`
integration-test:
	docker-compose -f integration-tests/docker-compose.yml up -d --build
	$(GOTEST) -v ./integration-tests/...
	docker-compose -f integration-tests/docker-compose.yml down
lint:
	$(GOLINT) ./...
clean:
	$(GOCLEAN)
	rm -r $(BINARY_DIR)
run:
	mkdir -p $(BINARY_DIR) && $(GOBUILD) -o $(BINARY_DIR) ./...
	./$(BINARY_DIR)/$(BINARY_NAME)