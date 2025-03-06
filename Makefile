PREFIX ?= /usr/local/bin
BINARY_NAME = hydectl
VERSION ?= $(shell git describe --tags --always --dirty)
GIT = github.com/HyDE-Project/hydectl/

all: uninstall clean build install

build:
	go build -ldflags "-X github.com/hyde-project/hydectl/cmd.Version=$(VERSION)" -o $(BINARY_NAME)

install: build
	@if [ "$$(id -u)" -eq 0 ]; then \
		echo "Installing to $(PREFIX)"; \
		install -Dm755 $(BINARY_NAME) $(PREFIX)/$(BINARY_NAME); \
	else \
		echo "Installing to $$HOME/.local/bin"; \
		mkdir -p $$HOME/.local/bin; \
		install -Dm755 $(BINARY_NAME) $$HOME/.local/bin/$(BINARY_NAME); \
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
	rm -f $(BINARY_NAME)

.PHONY: all build install uninstall completion clean
