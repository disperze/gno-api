#!/usr/bin/make -f

export GO111MODULE = on

BUILD_FLAGS := -ldflags '-w -s' -trimpath

all: install

install: go.sum
	go install -mod=readonly $(BUILD_FLAGS) ./cmd

build:
	go build $(BUILD_FLAGS) -o build/gnoapi ./cmd

.PHONY: all install build