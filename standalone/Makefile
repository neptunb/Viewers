.PHONY: all viewer launcher package clean deps tidy run

STANDALONE_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
LAUNCHER_DIR := $(STANDALONE_DIR)/launcher
DIST_DIR := $(STANDALONE_DIR)/dist

all: viewer package

deps:
	cd $(LAUNCHER_DIR) && go mod download

tidy:
	cd $(LAUNCHER_DIR) && go mod tidy

viewer:
	bash $(STANDALONE_DIR)/scripts/build-viewer.sh

launcher:
	cd $(LAUNCHER_DIR) && go build -o $(DIST_DIR)/viewer-dev ./cmd/viewer

package:
	bash $(STANDALONE_DIR)/scripts/package.sh

run: launcher
	$(DIST_DIR)/viewer-dev --study $(STANDALONE_DIR)/example-study

clean:
	rm -rf $(DIST_DIR)
	rm -rf $(LAUNCHER_DIR)/internal/webassets/web/*
	touch $(LAUNCHER_DIR)/internal/webassets/web/.gitkeep
