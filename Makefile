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
	docker compose up -d && \
	TESTDB="postgres://examine:examine@127.0.0.1:5432/examine?sslmode=disable" \
		go test -tags=dbtest -v ./... | grep -v '^?'

# go install github.com/mitranim/gow@latest
.PHONY: serve
serve:
	cd ${MKFILE_DIR}
	docker compose up -d && \
	gow -v -s \
		run -trimpath ${MKFILE_DIR}cmd/server/ \
			-dev=true \
			-logLevel=trace \
			-logPretty=true \
			-config=${MKFILE_DIR}configs/server/server.toml \
			-migrate=${MKFILE_DIR}configs/db_migrate/

.PHONY: db_cli
db_cli:
	cd ${MKFILE_DIR} && docker compose exec postgres psql -U examine -d examine

.PHONY: clean
clean:
	cd ${MKFILE_DIR} && docker compose down -v
	rm -rf ${DIST_DIR}/*

.PHONY: wire
wire:
	wire gen ./...

.PHONY: revive
revive:
	revive -config revive.toml -exclude ./vendor/... ./...
