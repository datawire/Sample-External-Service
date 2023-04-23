##@ General

define USAGE_OPTIONS

Options:

  \033[36mBINARIES\033[0m       
		 The binaries to build.
		 This option is available when using: make build|build-multiarch
		 Example: \033[36mmake build BINARIES="eg-ext"\033[0m
		 Default is all of services/ directory.
  \033[36mIMAGES\033[0m     
		 Backend images to make.
		 This option is available when using: make docker-build|docker-build-multiarch|docker-push|docker-push-multiarch
		 Example: \033[36mmake docker-build-multiarch IMAGES="eg-ext"\033[0m
		 Default is all of docker/ directory.
  \033[36mPLATFORM\033[0m   
		 The specified platform to build.
		 This option is available when using: make build
		 Example: \033[36mmake build BINARIES="eg-ext" PLATFORM="linux_amd64"\033[0m
		 Supported Platforms: linux_amd64 linux_arm64 darwin_amd64 darwin_arm64
		 Default is the platform constructured from system `go env`
  \033[36mPLATFORMS\033[0m  
		 The multiple platforms to build.
		 This option is available when using: make build-multiarch.
		 Example: \033[36mmake build-multiarch BINARIES="eg-ext" PLATFORMS="linux_amd64 linux_arm64"\033[0m
		 Default is "linux_amd64 linux_arm64 darwin_amd64 darwin_arm64"
  \033[36mREGISTRY\033[0m  
		 The image registry to use when building/pushing docker images.
		 This option is available when using: make docker-build|docker-build-multiarch|docker-push|docker-push-multiarch
		 Use docker.io/datawiredev for development and docker.io/datawire for releases.
		 Example: \033[36mmake docker-build-multiarch IMAGES="eg-ext" REGISTRY="myregistry"\033[0m
		 Default is docker.io/datawiredev
  \033[36mCHARTS\033[0m
		 Helm charts to package or lint.
		 This option is available when using: make helm-package|helm-lint.
		 Example: \033[36mmake helm-lint CHARTS="eg-plus"\033[0m
		 Default is all the charts found in the helm/ directory
endef
export USAGE_OPTIONS

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

# TODO: Alice needs updated
.PHONY: help
help: ## Display this help
	@echo -e "Envoy Gateway Plus is a paid-feature extension built on top of Envoy Gateway, adding a set of proprietary plugins and distributed with enterprise support\n"
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make \033[36m<target>\033[0m \033[36m<options>\033[0m\n\nTargets\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-25s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
	@echo -e "$$USAGE_OPTIONS"
