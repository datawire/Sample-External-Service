# This file contains a set common variables
#
# All make targets related to common variables are defined in this file.

# ==================================================================================================
# Root Settings:
# ==================================================================================================

# Set project root directory path
ifeq ($(origin ROOTDIR),undefined)
ROOTDIR := $(abspath $(shell  pwd -P))
endif

# Location to install binary builds to
LOCALBIN ?= $(ROOTDIR)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

# git HEAD commit SHA
REV := $(shell git rev-parse --short HEAD)

# ==================================================================================================
# Go Settings:
# ==================================================================================================

# Get the host GOOS, GOARCH from go env
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

# Verify that the system Go version is compatible with the go version found in go.mod
GO_MOD_GOVERSION := $(shell grep -E "go 1.[0-9]+.*[0-9]*" go.mod | awk '{print $$2}')
SYSTEM_GOVERSION := $(shell go version | awk '{ print $$3 }' | sed 's/go//')

MAJOR_GO_MOD_GOVERSION := $(word 1,$(subst ., , $(GO_MOD_GOVERSION)))
MAJOR_SYSTEM_GOVERSION := $(word 1,$(subst ., , $(SYSTEM_GOVERSION)))
MINOR_GO_MOD_GOVERSION := $(word 2,$(subst ., , $(GO_MOD_GOVERSION)))
MINOR_SYSTEM_GOVERSION := $(word 2,$(subst ., , $(SYSTEM_GOVERSION)))

IS_GO_VERSION_VALID := $(shell test $(MAJOR_SYSTEM_GOVERSION) -eq $(MAJOR_GO_MOD_GOVERSION) && \
                               test $(MINOR_SYSTEM_GOVERSION) -ge $(MINOR_GO_MOD_GOVERSION); \
                               echo $$?)

ifneq (${IS_GO_VERSION_VALID},0)
  $(error This Makefile requires Go $(GO_MOD_GOVERSION) or newer; you have $(SYSTEM_GOVERSION))
endif

# TODO: go-mkopensource requires downloading the golang source code tar in order to
# determine the version used for the standard library. Use the go version in go.mod as
# the canonical go version for builds and enforce across make that go version
# is installed on systems
#
# Go source code tar for generating DEPENDENCIES.md
GO_SRC ?= go$(SYSTEM_GOVERSION).src.tar.gz

