# EchoState — Build & Release
#
# Targets:
#   build-daemon          Build daemon for host platform
#   build-daemon-all      Cross-compile daemon for all platforms
#   build-api             Install API production dependencies
#   build-extension       Package Chrome extension as ZIP
#   build-all             Build everything
#   clean                 Remove build artifacts
#   release               Build all artifacts for release
#   test                  Run all tests (Go vet + compile check)
#   fmt                   Format all Go code

DAEMON_DIR  := daemon
API_DIR     := api
EXT_DIR     := extension
BIN_DIR     := bin
DAEMON_NAME := echostate-daemon

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -ldflags="-X main.version=$(VERSION) -X main.commit=$(COMMIT)"

.PHONY: build-daemon build-daemon-all build-api build-extension build-all clean release test fmt

# ------------------------------------------------------------------
# Go daemon — host platform
# ------------------------------------------------------------------
build-daemon:
	@mkdir -p $(BIN_DIR)
	cd $(DAEMON_DIR) && go build $(LDFLAGS) -o ../$(BIN_DIR)/$(DAEMON_NAME) .

# ------------------------------------------------------------------
# Go daemon — cross‑platform (all supported targets)
# ------------------------------------------------------------------
build-daemon-all:
	@mkdir -p $(BIN_DIR)
	@set -e; \
	for goos in linux darwin windows; do \
		for goarch in amd64 arm64; do \
			if [ "$$goos" = "windows" ] && [ "$$goarch" = "arm64" ]; then continue; fi; \
			ext=; \
			if [ "$$goos" = "windows" ]; then ext=".exe"; fi; \
			echo "  building $$goos/$$goarch ..."; \
			cd $(DAEMON_DIR) && GOOS=$$goos GOARCH=$$goarch go build $(LDFLAGS) \
				-o ../$(BIN_DIR)/$(DAEMON_NAME)-$$goos-$$goarch$$ext .; \
			cd ..; \
		done; \
	done

# ------------------------------------------------------------------
# Node API
# ------------------------------------------------------------------
build-api:
	cd $(API_DIR) && npm ci --production

# ------------------------------------------------------------------
# Chrome extension — plain source zip (no build step needed)
# ------------------------------------------------------------------
build-extension:
	@mkdir -p $(BIN_DIR)
	python3 -c "import zipfile, os; zf=zipfile.ZipFile('$(BIN_DIR)/echostate-extension-$(VERSION).zip','w'); [zf.write(os.path.join('$(EXT_DIR)',f),f) for f in ['manifest.json','content.js','style.css']]; zf.close()"

# ------------------------------------------------------------------
# All-in-one (host only)
# ------------------------------------------------------------------
build-all: build-daemon build-api build-extension

# ------------------------------------------------------------------
# Clean
# ------------------------------------------------------------------
clean:
	rm -rf $(BIN_DIR)

# ------------------------------------------------------------------
# Release — full cross-compile + packaging
# ------------------------------------------------------------------
release: build-daemon-all build-api build-extension
	@echo ""
	@echo "=== Release $(VERSION) artifacts ==="
	@ls -lh $(BIN_DIR)/

# ------------------------------------------------------------------
# Test (vet + compile)
# ------------------------------------------------------------------
test:
	cd $(DAEMON_DIR) && go vet ./... && echo "go vet: OK"
	cd $(API_DIR) && node --check index.js && echo "node check: OK"

# ------------------------------------------------------------------
# Format
# ------------------------------------------------------------------
fmt:
	cd $(DAEMON_DIR) && go fmt ./...
