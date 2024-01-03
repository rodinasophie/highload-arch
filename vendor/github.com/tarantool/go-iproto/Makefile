SHELL := /bin/bash

.PHONY: all
all: generate test

.PHONY: deps
deps:
	go install golang.org/x/tools/cmd/goimports@latest
	go install golang.org/x/tools/cmd/stringer@latest

.PHONY: format
format:
	goimports -l -w .

.PHONY: generate
generate:
	go generate ./...

.PHONY: test
test:
	@echo "Running all packages tests"
	go clean -testcache
	go test -tags ./... -v -p 1
