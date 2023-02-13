# suppress output, run `make XXX V=` to be verbose
V := @

# Semantic versioning format https://semver.org/
tag_regex := ^v([0-9]{1,}\.){2}[0-9]{1,}$

# Common
NAME = go.harvest
VCS = github.com
ORG = isaevdimka
VERSION ?= $(shell git describe --always --tags)
CURRENT_TIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

# Build
OUT_DIR = ./bin
MAIN_PKG = ./cmd/${NAME}
ACTION ?= build
GC_FLAGS = -gcflags 'all=-N -l'
LD_FLAGS = -ldflags "-s -v -w -X 'main.version=${VERSION}' -X 'main.buildTime=${CURRENT_TIME}'"

BUILD_CMD = CGO_ENABLED=1 go build -o ${OUT_DIR}/${NAME} ${LD_FLAGS} ${MAIN_PKG}
DEBUG_CMD = CGO_ENABLED=1 go build -o ${OUT_DIR}/${NAME} ${GC_FLAGS} ${MAIN_PKG}

# Other
.PHONY: help
.DEFAULT_GOAL := help
help:
	@grep -h -E '^[a-zA-Z\._-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build: ## build production
	@echo BUILDING PRODUCTION $(NAME)
	$(V)${BUILD_CMD}
	@echo DONE

.PHONY: build-debug
build-debug: ## build debug
	@echo BUILDING DEBUG $(NAME)
	$(V)${DEBUG_CMD}
	@echo DONE


.PHONY: clean
clean: ## cleanup build
	@echo "Removing $(OUT_DIR)"
	$(V)rm -rf $(OUT_DIR)
	@echo "Cleanup go modcache"
	$(V)GOPRIVATE=${VCS}/* go clean --modcache

.PHONY: vendor
vendor: ## bump vendor
	$(V)GOPRIVATE=${VCS}/* go mod tidy
	$(V)GOPRIVATE=${VCS}/* go mod vendor
	$(V)git add vendor go.mod go.sum

.PHONY: vendor
run: ## run
	$(V)./bin/go.harvest
