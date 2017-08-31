AGENTS=api scheduler aggregator worker
ROOT_DIR=github.com/ovh/metronome
SRC_DIR=$(ROOT_DIR)/src
BUILD_DIR=build

CC=go build
CFLAGS=-race

rwildcard=$(foreach d,$(wildcard $1*),$(call rwildcard,$d/,$2) $(filter $(subst *,%,$2),$d))
VPATH= $(BUILD_DIR)

.SECONDEXPANSION:

.PHONY: all
all: agents

.PHONY: agents
agents: assets $(AGENTS)

$(AGENTS): $$(call rwildcard, src/$$@, *.go) $$(call rwildcard, src/metronome, *.go)
	$(CC) $(CFLAGS) -o $(BUILD_DIR)/$@ $(SRC_DIR)/$@

.PHONY: install
install: agents
	@for a in $(AGENTS); do cp "$(BUILD_DIR)/$$a" "$$GOPATH/bin/metronome-$$a"; done

.PHONY: lint
lint:
	@command -v gometalinter >/dev/null 2>&1 || { echo >&2 "gometalinter is required but not available please follow instructions from https://github.com/alecthomas/gometalinter"; exit 1; }
	gometalinter --deadline=180s --disable-all --enable=gofmt ./src/...
	gometalinter --deadline=180s --disable-all --enable=vet ./src/...
	gometalinter --deadline=180s --disable-all --enable=golint ./src/...
	gometalinter --deadline=180s --disable-all --enable=ineffassign ./src/...
	gometalinter --deadline=180s --disable-all --enable=misspell ./src/...
	gometalinter --deadline=180s --disable-all --enable=staticcheck ./src/...

.PHONY: format
format:
	gofmt -w -s ./src

.PHONY: assets
assets: src/api/core/assets.go src/metronome/pg/schema.go

src/api/core/assets.go: $$(call rwildcard, src/api, *.json)
	@command -v go-bindata >/dev/null 2>&1 || { echo >&2 "go-bindata is required but not available please follow instructions from https://github.com/lestrrat/go-bindata"; exit 1; }
	go-bindata -ignore=\\.*\\.go -o src/api/core/assets.go -pkg core -prefix "src/api/controllers" src/api/...

src/metronome/pg/schema.go: $$(call rwildcard, src/metronome/pg/schema, *.sql)
	@command -v go-bindata >/dev/null 2>&1 || { echo >&2 "go-bindata is required but not available please follow instructions from https://github.com/lestrrat/go-bindata"; exit 1; }
	go-bindata -ignore=\\.*\\.go -o src/metronome/pg/schema.go -pkg pg -prefix "src/metronome/pg/schema" src/metronome/pg/schema/...


.PHONY: dev
dev: assets format lint agents

.PHONY: test
test:
	ginkgo -r ./src
.PHONY: testrun
testrun:
	ginkgo watch -r ./src

.PHONY: clean
clean:
	-rm -r build
