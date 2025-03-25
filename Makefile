PREFIX ?= /usr/local/bin
BINARY_NAME = hydectl
GIT_COUNT := $(shell git rev-list --count HEAD)
GIT_HASH := $(shell git rev-parse --short HEAD)
GIT_DESCRIBE := $(shell git describe --tags --always)
VERSION ?= r$(GIT_COUNT).$(GIT_HASH)



GIT = github.com/HyDE-Project/hydectl/

all: uninstall clean build install

build:
	@echo "Building $(BINARY_NAME) $(VERSION)"
	go build -ldflags "-X hydectl/cmd.Version=$(VERSION)" -o bin/$(BINARY_NAME)

install: build
	@if [ "$$(id -u)" -eq 0 ]; then \
		echo "Installing to $(PREFIX)"; \
		install -Dm755 bin/$(BINARY_NAME) $(PREFIX)/$(BINARY_NAME); \
	else \
		echo "Installing to $$HOME/.local/bin"; \
		mkdir -p $$HOME/.local/bin; \
		install -Dm755 bin/$(BINARY_NAME) $$HOME/.local/bin/$(BINARY_NAME); \
	fi

uninstall:
	@if [ "$$(id -u)" -eq 0 ]; then \
		echo "Removing from $(PREFIX)"; \
		rm -f $(PREFIX)/$(BINARY_NAME); \
	else \
		echo "Removing from $$HOME/.local/bin"; \
		rm -f $$HOME/.local/bin/$(BINARY_NAME); \
	fi

completion:
	@echo "Generating completion script"
	@$(BINARY_NAME) completion bash > /etc/bash_completion.d/$(BINARY_NAME)

clean:
	rm -f bin/$(BINARY_NAME)

.PHONY: all build install uninstall completion clean
