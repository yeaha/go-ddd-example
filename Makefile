SHELL:=/bin/sh

export GO111MODULE=on
export GOPROXY=https://goproxy.cn,direct

# Path Related
MKFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
MKFILE_DIR := $(dir $(MKFILE_PATH))
DIST_DIR := ${MKFILE_DIR}build/dist

# Rules
.PHONY: test
test:
	cd ${MKFILE_DIR}
	go test -v ./... | grep -v '^?'

.PHONY: alltest
alltest:
	cd ${MKFILE_DIR}
	go test -tags=dbtest -v ./... | grep -v '^?'

# go install github.com/mitranim/gow@latest
.PHONY: serve
serve:
	cd ${MKFILE_DIR}
	gow -v -s \
		run -trimpath ${MKFILE_DIR}cmd/server/ \
			-logLevel=trace \
			-dbDir=${MKFILE_DIR}db/ \
			-config=${MKFILE_DIR}configs/server/server.toml

.PHONY: clean
clean:
	rm -rf ${DIST_DIR}/*
	rm -rf ${MKFILE_DIR}db/*

.PHONY: wire
wire:
	wire gen ./...

.PHONY: revive
revive:
	revive -config revive.toml -exclude ./vendor/... ./...
