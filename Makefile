SHELL = bash
PROJECT_ROOT := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
THIS_OS := $(shell uname)

GOTESTSUM_VERSION ?= v1.6.2
GOLANGCI_LINT ?= $(shell which golangci-lint)
GOLANGCI_LINT_VERSION ?= v1.38.0

# Using directory as project name.
PROJECT_NAME := $(shell basename $(PROJECT_ROOT))
PROJECT_MODULE := $(shell go list -m)

GIT_COMMIT=$$(git rev-parse --short HEAD)
GIT_DIRTY=$$(test -n "`git status --porcelain`" && echo "+CHANGES" || true)
GIT_DESCRIBE=$$(git describe --tags --always --match "v*")

GIT_IMPORT := "$(PROJECT_MODULE)/internal/version"

GO_LDFLAGS := "-s -w -X $(GIT_IMPORT).GitCommit=$(GIT_COMMIT)$(GIT_DIRTY) -X $(GIT_IMPORT).GitDescribe=$(GIT_DESCRIBE)$(GIT_DIRTY)"

GO_TEST_CMD = $(if $(shell command -v gotestsum 2>/dev/null),gotestsum --,go test)
GO_TEST_PKGS ?= "./..."

default: help

ifeq ($(CI),true)
$(info Running in a CI environment, verbose mode is disabled)
else
VERBOSE="true"
endif

ALL_TARGETS = linux_386 \
	linux_amd64 \
	linux_arm \
	linux_arm64 \
	windows_386 \
	windows_amd64 \
	darwin_amd64

SUPPORTED_OSES = Darwin Linux FreeBSD Windows MSYS_NT

# include per-user customization after all variables are defined
-include Makefile.local

dist/%/authzy: GO_OUT ?= $@
dist/%/authzy: ## Build for GOOS_GOARCH, e.g. dist/linux_amd64/authzy
ifeq (,$(findstring $(THIS_OS),$(SUPPORTED_OSES)))
	$(warning WARNING: Building is only supported on $(SUPPORTED_OSES); not $(THIS_OS))
endif
	@echo "==> Building $@ ..."
	@CGO_ENABLED=0 \
		GOOS=$(firstword $(subst _, ,$*)) \
		GOARCH=$(lastword $(subst _, ,$*)) \
		go build -trimpath -ldflags $(GO_LDFLAGS) -o $(GO_OUT) ./cmd/$(PROJECT_NAME)/main.go

dist/windows_%/authzy: GO_OUT = $@.exe

# Define package targets for each of the build targets we actually have on this system
define makePackageTarget

dist/$(1).zip: dist/$(1)/$(PROJECT_NAME)
	@echo "==> Packaging for $(1)..."
	@zip -j dist/$(1).zip dist/$(1)/*
	@cat dist/$(1).zip | sha256sum > dist/$(1).zip.sha256
	@truncate -s 64 dist/$(1).zip.sha256

endef

# Reify the package targets
$(foreach t,$(ALL_TARGETS),$(eval $(call makePackageTarget,$(t))))

# Only for CI compliance
.PHONY: bootstrap
bootstrap: deps lint-deps # Install all dependencies

.PHONY: deps
deps:  ## Install build and development dependencies
	@echo "==> Updating build dependencies..."
	go install gotest.tools/gotestsum@$(GOTESTSUM_VERSION)

.PHONY: lint-deps
lint-deps: ## Install linter dependencies
	@echo "==> Updating linter dependencies..."
	@which golangci-lint 2>/dev/null || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin $(GOLANGCI_LINT_VERSION) && echo "Installed golangci-lint"

.PHONY: check
check: ## Lint the source code
	@echo "==> Linting source code..."
	$(GOLANGCI_LINT) run -j 1

	@echo "==> Checking Go mod..."
	$(MAKE) tidy
	@if (git status --porcelain | grep -Eq "go\.(mod|sum)"); then \
		echo go.mod or go.sum needs updating; \
		git --no-pager diff go.mod; \
		git --no-pager diff go.sum; \
		exit 1; fi

.PHONY: tidy
tidy:
	@echo "--> Tidy module"
	@go mod tidy

.PHONY: dev
dev: GOOS=$(shell go env GOOS)
dev: GOARCH=$(shell go env GOARCH)
dev: GOPATH=$(shell go env GOPATH)
dev: DEV_TARGET=dist/$(GOOS)_$(GOARCH)/$(PROJECT_NAME)
dev: ## Build for the current development platform
	@echo "==> Removing old development build..."
	@rm -f $(PROJECT_ROOT)/$(DEV_TARGET)
	@rm -f $(PROJECT_ROOT)/bin/$(PROJECT_NAME)
	@rm -f $(GOPATH)/bin/$(PROJECT_NAME)
	@$(MAKE) --no-print-directory \
		$(DEV_TARGET)
	@mkdir -p $(PROJECT_ROOT)/bin
	@mkdir -p $(GOPATH)/bin
	@cp $(PROJECT_ROOT)/$(DEV_TARGET) $(PROJECT_ROOT)/bin/
	@cp $(PROJECT_ROOT)/$(DEV_TARGET) $(GOPATH)/bin

.PHONY: install
install: GOOS=$(shell go env GOOS)
install: GOARCH=$(shell go env GOARCH)
install: GOPATH=$(shell go env GOPATH)
install: TARGET=dist/$(GOOS)_$(GOARCH)/$(PROJECT_NAME)
install: ## Install binary
	@echo "==> Removing old build..."
	@rm -f $(PROJECT_ROOT)/$(TARGET)
	@rm -f $(GOPATH)/bin/$(PROJECT_NAME)
	@$(MAKE) --no-print-directory \
		$(TARGET)
	@mkdir -p $(GOPATH)/bin
	@cp $(PROJECT_ROOT)/$(TARGET) $(GOPATH)/bin

.PHONY: build
build: clean $(foreach t,$(ALL_TARGETS),dist/$(t).zip) ## Build release packages
	@echo "==> Results:"
	@tree --dirsfirst $(PROJECT_ROOT)/dist

.PHONY: test
test: ## Run the test suite and/or any other tests
	@echo "==> Running test suites..."
	$(if $(ENABLE_RACE),GORACE="strip_path_prefix=$(GOPATH)/src") $(GO_TEST_CMD) \
		$(if $(ENABLE_RACE),-race) $(if $(VERBOSE),-v) \
		-cover \
		-coverprofile=coverage.out \
		-covermode=atomic \
		-timeout=15m \
		$(GO_TEST_PKGS)

.PHONY: coverage
coverage: ## Open a web browser displaying coverage
	go tool cover -html=coverage.out

.PHONY: clean
clean: ## Remove build artifacts
	@echo "==> Cleaning build artifacts..."
	@rm -fv coverage.out
	@find . -name '*.test' | xargs rm -fv
	@rm -rfv "$(PROJECT_ROOT)/bin/"
	@rm -rfv "$(PROJECT_ROOT)/dist/"
	@rm -fv "$(GOPATH)/bin/$(PROJECT_NAME)"

HELP_FORMAT="    \033[36m%-15s\033[0m %s\n"
.PHONY: help
help: ## Display this usage information
	@echo "Valid targets:"
	@grep -E '^[^ ]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		sort | \
		awk 'BEGIN {FS = ":.*?## "}; \
			{printf $(HELP_FORMAT), $$1, $$2}'
	@echo

FORCE:
