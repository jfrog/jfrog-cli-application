SHELL := /bin/bash
.DEFAULT_GOAL = build
GOCMD = go
export PROJECT_DIR ?= $(CURDIR)
BINARY_CLI = bin
WORKSPACE_ROOT = $(shell cd "${PROJECT_DIR}" && pwd)
TOOLS_DIR := $(CURDIR)/.tools
SCRIPTS_DIR = ${PROJECT_DIR}/scripts
TARGET_DIR = ${PROJECT_DIR}/target
LINKERFLAGS = -s -w
COMPILERFLAGS = all=-trimpath=$(WORKSPACE_ROOT)
GOOS = $(shell go env GOOS)
GOARCH = $(shell go env GOARCH)
GO_SOURCES = $(eval GO_SOURCES := $$(shell find . -type f -name "*.go" | grep -v ".*_mock\.go"))$(GO_SOURCES)

########## FORMAT ##########

format: GOFUMPT GOIMPORTS
	@${GOFUMPT} -w $(GO_SOURCES)
	@${GOIMPORTS} -w -local jfrog.com $(GO_SOURCES)

GOFUMPT = ${TOOLS_DIR}/gofumpt
GOFUMPT_VERSION = 0.5.0

GOFUMPT:
	${GOFUMPT} --version 2>/dev/null | grep ${GOFUMPT_VERSION} || GOBIN=${TOOLS_DIR} $(GOCMD) install mvdan.cc/gofumpt@v${GOFUMPT_VERSION}

GOIMPORTS = ${TOOLS_DIR}/goimports
GOIMPORTS_VERSION = 0.16.1

GOIMPORTS:
	GOBIN=${TOOLS_DIR} $(GOCMD) install golang.org/x/tools/cmd/goimports@v${GOIMPORTS_VERSION}

########## ANALYSE ##########

GOLANGCI_LINT         = ${TOOLS_DIR}/golangci-lint
GOLANGCI_LINT_VERSION = 1.63.4

verify: GOLANGCI_LINT
	echo $(GO_SOURCES)
	$(GOLANGCI_LINT) run --out-format tab --config "${WORKSPACE_ROOT}/.golangci.yml"

GOLANGCI_LINT:
	${GOLANGCI_LINT} --version 2>/dev/null | grep ${GOLANGCI_LINT_VERSION} || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ${TOOLS_DIR} v${GOLANGCI_LINT_VERSION}

########## BUILD ##########
prereq::
	$(GOCMD) install gotest.tools/gotestsum@latest
	GOBIN=${TOOLS_DIR} $(GOCMD) install go.uber.org/mock/mockgen@v0.6.0
	${TOOLS_DIR}/mockgen --version

build:: clean generate-mock
	$(GOCMD) env GOOS GOARCH
	$(GOCMD) build -ldflags="${LINKERFLAGS}" -gcflags ${COMPILERFLAGS} -o ${BINARY_CLI}/application-cli-plugin main.go


build-install:: build
	mkdir -p "${HOME}/.jfrog/plugins/application/bin"
	mv ${BINARY_CLI}/application-cli-plugin "${HOME}/.jfrog/plugins/application/bin/application"
	chmod +x "${HOME}/.jfrog/plugins/application/bin/application"

########## TEST ##########

.PHONY: clean-mock
clean-mock:
	@echo Cleaning generated mock files
	@find . -name "*_mock.go" -delete

.PHONY: clean
clean:: clean-mock
	@echo Cleaning generated files
	@rm -rf ${BINARY_CLI}

.PHONY: generate-mock
generate-mock: prereq clean-mock
	@echo Generating test mocks
	TOOLS_DIR=$(TOOLS_DIR) go generate ./...

test-prereq: generate-mock

test: PACKAGES=./...
test: test-prereq
	go test ./...
test-ci: test-prereq
	gotestsum --format testname --junitfile=utests-report.xml -- ./...
e2e-test: test-prereq
	go test ./e2e/... -tags=e2e
e2e-test-ci: test-prereq
	gotestsum --format testname --junitfile=e2e-tests-report.xml -- ./e2e/... -tags=e2e
