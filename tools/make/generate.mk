##@ Build Dependencies

# Location to install build tools & dependencies
TOOLSBIN := $(ROOTDIR)/tools/bin
$(TOOLSBIN):
	mkdir -p $(TOOLSBIN)

# ====================================================================================================
# Tool Versions:
# ====================================================================================================

GO_MKOPENSOURCE := $(TOOLSBIN)/go-mkopensource
GO_MKOPENSOURCE_VERSION ?= v0.0.7

# ====================================================================================================
# Installation & Binary Targets:
# ====================================================================================================

.PHONY: deps
deps: go-mkopensource ## Download all build dependencies

.PHONY: go-mkopensource
go-mkopensource: $(GO_MKOPENSOURCE) ## Download go-mkopensource if necessary
$(GO_MKOPENSOURCE): $(TOOLSBIN)
	test -x $(TOOLSBIN)/go-mkopensource || GOBIN=$(TOOLSBIN) go install github.com/datawire/go-mkopensource/cmd/go-mkopensource@$(GO_MKOPENSOURCE_VERSION)

.PHONY: generate
generate-deps: go-mkopensource ## generate DEPENDENCIES.md
## TODO: there's probably a better way but for now...
	curl -Ss https://dl.google.com/go/$(GO_SRC) -o /tmp/$(GO_SRC)
	$(GO_MKOPENSOURCE) --gotar /tmp/$(GO_SRC) --package=mod --output-format=txt --output-type=markdown > DEPENDENCIES.md
	rm -f /tmp/$(GO_SRC)

.PHONY: clean
clean: ## Clean up local bin, dist, gen, and testprofile directories
	rm -rf $(LOCALBIN)
