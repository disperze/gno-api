#!/usr/bin/make -f

export GO111MODULE = on

BUILD_FLAGS := -ldflags '-w -s' -trimpath

all: install

install: go.sum
	go install -mod=readonly $(BUILD_FLAGS) ./cmd

build:
	go build -o build/gnoapi ./cmd

build-indexer:
	go build -o build/gnoind ./indexer

.PHONY: all install build