# All make targets should be implemented in tools/make/*.mk
# ==================================================================================================
# Supported Targets: (Run `make help` to see more information)
# ==================================================================================================

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

# Here so that running just `make` will run `make build`
.PHONY: all
all: build

include tools/make/globals.mk
include tools/make/help.mk
include tools/make/generate.mk
include tools/make/lint.mk
include tools/make/build.mk
