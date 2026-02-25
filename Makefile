BINARY_NAME := jcompressor
CMD_PATH := ./cmd/jcompressor
BUILD_DIR := ./build

PREFIX ?= /usr/local
INSTALL_DIR ?= $(PREFIX)/bin

GO := go
GOOS := $(shell $(GO) env GOOS)
GOARCH := $(shell $(GO) env GOARCH)
GOVERSION := $(shell $(GO) version)

.PHONY: all build env install uninstall clean help

all: build

env:
	@echo "Go environment summary:"
	@echo "  GOOS:    $(GOOS)"
	@echo "  GOARCH:  $(GOARCH)"
	@echo "  go:      $(GOVERSION)"
	@echo "  Module:  $(shell awk '/^module /{print $$2}' go.mod)"
	@echo "  Binary:  $(BINARY_NAME) (built from $(CMD_PATH))"

# Собрать бинарник в $(BUILD_DIR)
build:
	@echo "Building $(BINARY_NAME) for $(GOOS)/$(GOARCH) (without CGO)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_PATH)
	@echo "Built: $(BUILD_DIR)/$(BINARY_NAME)"
	@echo "Note: WebP support disabled (CGO_ENABLED=0). To enable, use 'make build-webp'."

# Собрать бинарник с поддержкой WebP (требует CGO и libwebp)
build-webp:
	@echo "Building $(BINARY_NAME) for $(GOOS)/$(GOARCH) with WebP support (CGO enabled)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=1 GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_PATH)
	@echo "Built: $(BUILD_DIR)/$(BINARY_NAME) (with WebP support)"

# Install the built binary to $(INSTALL_DIR). Uses sudo if necessary.
install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_DIR)"
	@mkdir -p $(INSTALL_DIR)
	@if [ -w $(INSTALL_DIR) ]; then \
		install -m 0755 $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME); \
		echo "Installed to $(INSTALL_DIR)/$(BINARY_NAME)"; \
	else \
		echo "Requires sudo to install to $(INSTALL_DIR)"; \
		sudo install -m 0755 $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME); \
		echo "Installed to $(INSTALL_DIR)/$(BINARY_NAME)"; \
	fi

# Uninstall the binary from $(INSTALL_DIR)
uninstall:
	@echo "Uninstalling $(INSTALL_DIR)/$(BINARY_NAME)"
	@if [ -e $(INSTALL_DIR)/$(BINARY_NAME) ]; then \
		if [ -w $(INSTALL_DIR)/$(BINARY_NAME) ]; then \
			rm -f $(INSTALL_DIR)/$(BINARY_NAME); \
			echo "Removed $(INSTALL_DIR)/$(BINARY_NAME)"; \
		else \
			echo "Requires sudo to remove $(INSTALL_DIR)/$(BINARY_NAME)"; \
			sudo rm -f $(INSTALL_DIR)/$(BINARY_NAME); \
			echo "Removed $(INSTALL_DIR)/$(BINARY_NAME)"; \
		fi \
	else \
		echo "Nothing to do: $(INSTALL_DIR)/$(BINARY_NAME) not found"; \
	fi

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)

help:
	@echo "Usage: make [target]"
	@echo "Targets:"
	@echo "  all (default)   - same as 'build'"
	@echo "  env             - print Go build environment and module info"
	@echo "  build           - build the binary into $(BUILD_DIR) (without WebP)"
	@echo "  build-webp      - build with WebP support (requires CGO and libwebp)"
	@echo "  install         - install the binary into \\$(INSTALL_DIR) (uses sudo if needed)"
	@echo "                   Override PREFIX to change location, e.g. 'make install PREFIX=/usr'"
	@echo "  uninstall       - remove the installed binary from \\$(INSTALL_DIR)"
	@echo "  clean           - remove build artifacts"
