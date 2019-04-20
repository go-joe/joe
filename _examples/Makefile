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
	@gofmt -d "$@"
	@test -z "$$(gofmt -l "$@")" || (echo "ERROR: bad source code formatting in $@" && false)
	cd "$(dir $@)" && go build -o /dev/null -v
