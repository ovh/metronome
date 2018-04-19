# Variables
BUILD_DIR 		:= build
GITHASH 			:= $(shell git rev-parse HEAD)
VERSION				:= $(shell git describe --abbrev=0 --tags --always)
DATE					:= $(shell TZ=UTC date -u '+%Y-%m-%dT%H:%M:%SZ UTC')
LINT_PATHS		:= ./src/...
FORMAT_PATHS 	:= ./src

# Compilation variables
CC 						:= go build
DFLAGS 				:= -race
CFLAGS 				:= -X 'github.com/ovh/metronome/src/aggregator/cmd.githash=$(GITHASH)' \
	-X 'github.com/ovh/metronome/src/aggregator/cmd.date=$(DATE)' \
	-X 'github.com/ovh/metronome/src/aggregator/cmd.version=$(VERSION)' \
	-X 'github.com/ovh/metronome/src/api/cmd.githash=$(GITHASH)' \
	-X 'github.com/ovh/metronome/src/api/cmd.date=$(DATE)' \
	-X 'github.com/ovh/metronome/src/api/cmd.version=$(VERSION)' \
	-X 'github.com/ovh/metronome/src/scheduler/cmd.githash=$(GITHASH)' \
	-X 'github.com/ovh/metronome/src/scheduler/cmd.date=$(DATE)' \
	-X 'github.com/ovh/metronome/src/scheduler/cmd.version=$(VERSION)' \
	-X 'github.com/ovh/metronome/src/worker/cmd.githash=$(GITHASH)' \
	-X 'github.com/ovh/metronome/src/worker/cmd.date=$(DATE)' \
	-X 'github.com/ovh/metronome/src/worker/cmd.version=$(VERSION)'
CROSS					:= GOOS=linux GOARCH=amd64

# Makefile variables
VPATH 				:= $(BUILD_DIR)

# Function definitions
rwildcard			:= $(foreach d,$(wildcard $1*),$(call rwildcard,$d/,$2) $(filter $(subst *,%,$2),$d))

.SECONDEXPANSION:
.PHONY: all
all: init dep format lint release

.PHONY: init
init:
	go get -u github.com/golang/dep/...
	go get -u github.com/alecthomas/gometalinter
	go get -u github.com/gobuffalo/packr/...
	go get -u github.com/onsi/ginkgo/ginkgo
	go get -u golang.org/x/tools/cmd/cover
	go get -u github.com/modocache/gover
	$(GOPATH)/bin/gometalinter --install --no-vendored-linters

.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)
	rm -rf dist
	$(GOPATH)/bin/packr clean -v

.PHONY: dep
dep: assets
	$(GOPATH)/bin/dep ensure -v

.PHONY: format
format:
	gofmt -w -s $(FORMAT_PATHS)

.PHONY: lint
lint:
	$(GOPATH)/bin/gometalinter --disable-all --config .gometalinter.json $(LINT_PATHS)

.PHONY: test
test:
	$(GOPATH)/bin/ginkgo -r --randomizeAllSpecs --randomizeSuites --failOnPending --cover --trace --race --progress --compilers=2

.PHONY: testrun
testrun:
	$(GOPATH)/bin/ginkgo watch -r ./src

.PHONY: cover
cover:
	$(GOPATH)/bin/gover ./src coverage.txt

.PHONY: assets
assets: src/api/core/core-packr.go src/metronome/pg/pg-packr.go

%-packr.go:
	$(GOPATH)/bin/packr -v

.PHONY: dev
dev: format lint build

.PHONY: build
build: assets $$(call rwildcard, ./src/aggregator, *.go) $$(call rwildcard, ./src/api, *.go) $$(call rwildcard, ./src/worker, *.go) $$(call rwildcard, ./src/scheduler, *.go) $$(call rwildcard, ./src/metronome, *.go)
	$(CC) $(DFLAGS) -ldflags "-s -w $(CFLAGS)" -o $(BUILD_DIR)/aggregator src/aggregator/aggregator.go
	$(CC) $(DFLAGS) -ldflags "-s -w $(CFLAGS)" -o $(BUILD_DIR)/api src/api/api.go
	$(CC) $(DFLAGS) -ldflags "-s -w $(CFLAGS)" -o $(BUILD_DIR)/scheduler src/scheduler/scheduler.go
	$(CC) $(DFLAGS) -ldflags "-s -w $(CFLAGS)" -o $(BUILD_DIR)/worker src/worker/worker.go

.PHONY: release
release: assets $$(call rwildcard, ./src/aggregator, *.go) $$(call rwildcard, ./src/api, *.go) $$(call rwildcard, ./src/worker, *.go) $$(call rwildcard, ./src/scheduler, *.go) $$(call rwildcard, ./src/metronome, *.go)
	$(CC) -ldflags "-s -w $(CFLAGS)" -o $(BUILD_DIR)/aggregator src/aggregator/aggregator.go
	$(CC) -ldflags "-s -w $(CFLAGS)" -o $(BUILD_DIR)/api src/api/api.go
	$(CC) -ldflags "-s -w $(CFLAGS)" -o $(BUILD_DIR)/scheduler src/scheduler/scheduler.go
	$(CC) -ldflags "-s -w $(CFLAGS)" -o $(BUILD_DIR)/worker src/worker/worker.go

.PHONY: dist
dist: assets $$(call rwildcard, ./src/aggregator, *.go) $$(call rwildcard, ./src/api, *.go) $$(call rwildcard, ./src/worker, *.go) $$(call rwildcard, ./src/scheduler, *.go) $$(call rwildcard, ./src/metronome, *.go)
	$(CROSS) $(CC) -ldflags "-s -w $(CFLAGS)" -o $(BUILD_DIR)/aggregator src/aggregator/aggregator.go
	$(CROSS) $(CC) -ldflags "-s -w $(CFLAGS)" -o $(BUILD_DIR)/api src/api/api.go
	$(CROSS) $(CC) -ldflags "-s -w $(CFLAGS)" -o $(BUILD_DIR)/scheduler src/scheduler/scheduler.go
	$(CROSS) $(CC) -ldflags "-s -w $(CFLAGS)" -o $(BUILD_DIR)/worker src/worker/worker.go

.PHONY: install
install: release
	cp -v $(BUILD_DIR)/aggregator $(GOPATH)/bin/metronome-aggregator
	cp -v $(BUILD_DIR)/api $(GOPATH)/bin/metronome-api
	cp -v $(BUILD_DIR)/scheduler $(GOPATH)/bin/metronome-scheduler
	cp -v $(BUILD_DIR)/worker $(GOPATH)/bin/metronome-worker
