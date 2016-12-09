AGENTS=api scheduler aggregator worker
ROOT_DIR=github.com/runabove/metronome
SRC_DIR=$(ROOT_DIR)/src
BUILD_DIR=build

CC=go build
CFLAGS=

rwildcard=$(foreach d,$(wildcard $1*),$(call rwildcard,$d/,$2) $(filter $(subst *,%,$2),$d))
VPATH= $(BUILD_DIR)

.SECONDEXPANSION:

.PHONY : all
all : agents

.PHONY : agents
agents: assets $(AGENTS)

$(AGENTS): $$(call rwildcard, src/$$@, *.go) $$(call rwildcard, src/metronome, *.go)
	$(CC) $(CFLAGS) -o $(BUILD_DIR)/$@ $(SRC_DIR)/$@


.PHONY : assets
assets: src/api/core/assets.go

src/api/core/assets.go: $$(call rwildcard, src/api, *.json)
	go-bindata -ignore=\\.*\\.go -o src/api/core/assets.go -pkg core -prefix "src/api/controllers" src/api/...

.PHONY : clean
clean :
	-rm -r build
