# Define Go compiler and build flags
GOCMD=go
GOBUILD=$(GOCMD) build
BINARY_DIR=.
BINARY_BRANCH=server
BINARY_CLIENT=client

# Default target executed when no arguments are given to make.
all: build_branch build_client

build_branch:
	$(GOBUILD) -o $(BINARY_DIR)/$(BINARY_BRANCH) ./cmd/branch

build_client:
	$(GOBUILD) -o $(BINARY_DIR)/$(BINARY_CLIENT) ./cmd/client

clean:
	rm -f $(BINARY_DIR)/$(BINARY_BRANCH)
	rm -f $(BINARY_DIR)/$(BINARY_CLIENT)

.PHONY: all build_branch build_client clean
