SHELL := /bin/bash

EXAMPLES = $(shell find . -name 'main.go' | sort)
.PHONY: all $(EXAMPLES)

# all makes sure all examples actually compile and then updates the README.md
all: $(EXAMPLES)
	embedmd -w ../README.md

# check makes sure all examples actually compile and the README.md is in sync
check: $(EXAMPLES)
	@echo "embedmd -d ../README.md" && test -z "$$(embedmd -d ../README.md)" || (echo -e "\nERROR: Examples in README.md are out of sync!\nPlease run \"embedmd -w $(abspath ../README.md)\"" && false)

# We are building the files simply to detect any compile time issues.
*/main.go:
	@gofmt -d "$(@D)"
	@test -z "$$(gofmt -l "$(@D)")" || (echo "ERROR: bad source code formatting in $(@D)" && false)
	cd "$(dir $@)" && go build -o /dev/null -v
	cd "$(dir $@)" && go mod tidy
