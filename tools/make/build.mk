##@ Build

# List of building targets
SERVICES := $(wildcard ${ROOTDIR}/services/*)
BINARIES ?= $(foreach services,${SERVICES},$(notdir ${services}))

ifeq (${SERVICES},)
  $(error Could not determine SERVICES, set ROOTDIR or run in source dir)
endif
ifeq (${BINARIES},)
  $(error Could not determine BINARIES, set ROOTDIR or run in source dir)
endif

# Set a specific platform {os}_{arch} to build a binary for.
# If this is not set then it uses GOOS & GOARCH found on the system machine (see Go Setting section)
PLATFORM ?=
# If PLATFORM is not set by the user, set it using GOOS & GOARCH
ifeq ($(PLATFORM),)
PLATFORM = $(GOOS)_$(GOARCH)
endif

# Supported platforms for building multiarch binaries.
PLATFORMS ?= darwin_amd64 darwin_arm64 linux_amd64 linux_arm64

# Use docker.io/datawiredev as default for development.
REGISTRY ?=
# If PLATFORM is not set by the user, set it using GOOS & GOARCH
ifeq ($(REGISTRY),)
REGISTRY = docker.io/datawiredev
endif

# List of Docker images to build
DOCKER_DIRS := $(wildcard ${ROOTDIR}/docker/*)
IMAGES ?= $(foreach dir,${DOCKER_DIRS},$(notdir ${dir}))

# TAG is the tag to use when building/pushing image targets.
# Use the HEAD commit SHA as the default.
TAG ?= $(REV)

# Supported Platforms for building docker images with buildx.
IMAGE_PLATFORMS ?= linux_amd64 linux_arm64

# Convert IMAGE_PLATFORMS to buildx compatible format (e.g linux_amd64 linux_arm64 -> linux/amd64,linux/arm64),
BUILDX_PLATFORMS := $(shell echo "${IMAGE_PLATFORMS}" | sed "s|_|/|g;s| |,|g")

# Buildx context to use when building multiarch Docker images
BUILDX_CONTEXT ?= "ambassador-multiauth-builder"

# ====================================================================================================
# Go Binary Builds:
# ====================================================================================================

# Build the target binary in target platform.
# This is an intermediate target and not meant for public use.
# The pattern of _build.% is `_build.{platform}.{binary}` where platform={os}_{arch} e.g _build.linux_amd64.eg-ext
.PHONY: binary.%
binary.%: 
	$(eval BINARY := $(word 2,$(subst ., , $*)))
	$(eval PLATFORM_ARCH := $(word 1,$(subst ., ,$*)))
	$(eval OS := $(word 1,$(subst _, ,$(PLATFORM_ARCH))))
	$(eval ARCH := $(word 2,$(subst _, ,$(PLATFORM_ARCH))))
	CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) go build -o $(LOCALBIN)/$(OS)/$(ARCH)/$(BINARY) ./services/$(BINARY)

.PHONY: binaries
binaries: ## Build all of the binaries found in the services/ directory for all supported platforms. See PLATFORMS and BINARIES in Options for more info
binaries: $(foreach platform,$(PLATFORMS),$(addprefix binary.$(platform)., $(BINARIES)))

# ====================================================================================================
# Docker Builds:
# ====================================================================================================

.PHONY: _docker.verify
_docker.verify:
	$(eval ISDOCKER := $(shell docker version | grep "Engine" ))
	@if [ -z "$(ISDOCKER)" ]; then \
		echo "Cannot find docker, please install first"; \
		exit 1; \
	fi

.PHONY: _docker-buildx.verify
_docker-buildx.verify:
	$(eval ISBUILDX := $(shell docker buildx version | grep "docker/buildx" ))
	@if [ -z "$(ISBUILDX)" ]; then \
		echo "Cannot find docker buildx, please install first"; \
		exit 1; \
	fi

# Build docker image in target platform.
# The pattern of image.% is image.{platform}.{image} where platform={os}_{arch} e.g image.linux_amd64.eg-ext
image.%: _docker.verify binary.%
	$(eval IMAGE := $(word 2,$(subst ., , $*)))
	$(eval PLATFORM_ARCH := $(word 1,$(subst ., ,$*)))
	$(eval OS := $(word 1,$(subst _, ,$(PLATFORM_ARCH))))
	$(eval ARCH := $(word 2,$(subst _, ,$(PLATFORM_ARCH))))
	docker build --platform $(OS)/$(ARCH) --build-arg TARGETPLATFORM=$(OS)/$(ARCH) -f docker/$(IMAGE)/Dockerfile -t ${REGISTRY}/${IMAGE}:${TAG} $(LOCALBIN)

.PHONY: images
images: ## Build linux/amd64 docker images. See IMAGES in Options for more info
images: $(addprefix image.linux_amd64., $(IMAGES))

.PHONY: image.%
image.%:
	docker push ${REGISTRY}/$*:${TAG}

.PHONY: push
push: $(addprefix push., $(IMAGES)) ## Push docker images. See IMAGES and REGISTRY in Options for more info


# Build Docker image for all supporting platforms.
# The pattern of image-multiarch.% is _docker-build.{image} e.g. image-multiarch.multiauth
.PHONY: multiarch-image.%
multiarch-image.%: _docker-buildx.verify $(foreach platform,$(IMAGE_PLATFORMS),$(addprefix binary.$(platform).,%))
	-docker buildx rm $(BUILDX_CONTEXT)
	docker buildx create --use --name $(BUILDX_CONTEXT) --platform "${BUILDX_PLATFORMS}"
	docker buildx build $(LOCALBIN) -t ${REGISTRY}/$*:${TAG} -f docker/$*/Dockerfile --platform "${BUILDX_PLATFORMS}"

.PHONY: multiarch-images
multiarch-images: ## Build docker images for all supporting platforms. Currently that is linux/amd64 and linux/arm64. See IMAGES in Options for more info
multiarch-images: $(addprefix multiarch-image.,$(IMAGES))

# Build and push Docker images for all supporting platforms.
# The pattern of multiarch-push.% is multiarch-push.{image} e.g. multiarch-push.eg-ext
.PHONY: multiarch-push.%
_docker-push.%: _docker-buildx.verify $(foreach platform,$(IMAGE_PLATFORMS),$(addprefix binary.$(platform).,%))
	-docker buildx rm $(BUILDX_CONTEXT)
	docker buildx create --use --name $(BUILDX_CONTEXT) --platform "${BUILDX_PLATFORMS}"
	docker buildx build $(LOCALBIN) -t ${REGISTRY}/$*:${TAG} -f docker/$*/Dockerfile --platform "${BUILDX_PLATFORMS}" --push

.PHONY: multiarch-push
docker-push: ## Build and push docker images for all supporting platforms. See IMAGES and REGISTRY in Options for more info
docker-push: $(addprefix multiarch-push.,$(IMAGES))
