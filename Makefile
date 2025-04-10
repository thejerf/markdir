
BUILD_VERSION   ?= $(shell git describe --tags --always | sed 's/-/+/')
BUILD_TIMESTAMP := $(shell date -u '+%Y-%m-%d %H:%M:%SZ')
LDFLAGS         := -X "main.BuildVersion=$(BUILD_VERSION)" -X "main.BuildTime=$(BUILD_TIMESTAMP)"

.SILENT:

.PHONY: install
install:
	go install -ldflags '$(LDFLAGS)'
