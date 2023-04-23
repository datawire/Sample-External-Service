##@ Lint

.PHONY: lint
lint: go-fmt go-vet helm-lint ## Run all linters

# ====================================================================================================
# Go Linters & Formatters:
# ====================================================================================================

.PHONY: fmt
go-fmt: ## Run go fmt against code
	go fmt ./services/...

.PHONY: vet
go-vet: proto-gen ## Run go vet against code
	go vet ./services/...
