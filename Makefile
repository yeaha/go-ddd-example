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
	cd ${MKFILE_DIR} && go tool ginkgo -r -p --randomize-all --randomize-suites --keep-going --no-color ./internal/

.PHONY: test-e2e
test-e2e:
	cd ${MKFILE_DIR} && go tool ginkgo -r -p --randomize-all --randomize-suites --keep-going --timeout=2m --succinct --no-color ./test/e2e/

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

.PHONY: revive
revive:
	revive -config revive.toml -exclude ./vendor/... ./...
