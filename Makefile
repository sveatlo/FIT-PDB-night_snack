-include ./sedder_helper.mk

CMDS := snacker
ENVS := prod dev

CC = go
BIN_DIR = ./bin

CGO_ENABLED ?= 0
OS ?= linux
ARCH ?= amd64

BUILDIMAGE = golang:1.17-buster
RUNIMAGE = alpine
REGISTRY ?= ghcr.io/sveatlo/night_snack

# project_name is defined as the last part of package/git repo
# PROJECT_NAME = $(shell echo '$(GO_PROJECT_PACKAGE)' | awk -F'/' '{print $$NF}' )
PROJECT_NAME = night_snack
CMD_PKG_PATH = ./cmd/$(CMD_NAME)
CMD_CONFIG_FILENAME ?= $(CMD_NAME).yml
CMD_BIN_FILE = $(CMD_NAME)
CMD_BIN_PATH = $(BIN_DIR)/$(CMD_BIN_FILE)
CMD_IMAGE ?= $(REGISTRY)/$(CMD_NAME):$(VERSION)

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null | sed -e 's/^v//')
DATE = $(shell date +'%a, %d %b %Y %T %z')

GO_PROJECT_PACKAGE := $(shell cat go.mod | grep -E '^module' | awk '{print $$2}')
GO_LD_FLAGS = -X main.Version=${VERSION}
GO_CC_FLAGS = -ldflags "$(GO_LD_FLAGS)"
GO_CC_FLAGS_ESCAPED = $(shell echo '$(GO_CC_FLAGS)' | sed 's/"/\\\\"/g')

DOCS_DIR = ./docs

all: build

dev: SERVICES?=$(CMDS)
dev: dockerfile-dev
	docker-compose up --build $(SERVICES) cockroach mongo nats

build: $(BIN_DIR) $(foreach cmd,$(CMDS),build-$(cmd))

package: $(BIN_DIR) $(foreach cmd,$(CMDS),package-$(cmd))

.PHONY: proto
proto:
	protoc --proto_path=. -I ./proto -I $$GOPATH/src \
		--go_out=. \
		--go_opt=module=$(GO_PROJECT_PACKAGE) \
		proto/errors/*.proto
	protoc --proto_path=. -I ./proto -I $$GOPATH/src \
		--go_out=paths=source_relative:. \
		--go-grpc_out=paths=source_relative:. \
		proto/snacker/*.proto
	protoc --proto_path=. -I ./proto -I $$GOPATH/src \
		--go_out=paths=source_relative:. \
		--go-grpc_out=paths=source_relative:. \
		proto/restaurant/*.proto

define BUILD_template =
.PHONY: build-$(1)
build-$(1): CMD_NAME=$(1)
build-$(1):
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(OS) GOARCH=$(ARCH) $(CC) build $$(GO_CC_FLAGS) -o $$(CMD_BIN_PATH) $$(CMD_PKG_PATH)
endef
$(foreach cmd, $(CMDS), $(eval $(call BUILD_template,$(cmd))))

# 1 - cmd
# do not add build-$(cmd) as dependecy
# because this should be run in "debian" docker image without any `go` or `git` binaries
#  => `make build` should be run manually beforehand
define PACKAGE_template =
.PHONY: package-$(1)
package-$(1): CMD_NAME=$(1)
package-$(1):
	@echo ">>> Preparing environment"
	rm -rf $(1)_pkg_build
	cp -rp deployment/$(1)_pkg $(1)_pkg_build
	@echo "================================="
	@echo ">>> Preparing package content"
	mkdir -p $(1)_pkg_build/opt/moderntv/$(PROJECT_NAME)/bin
	cp $$(CMD_BIN_PATH) $(1)_pkg_build/opt/moderntv/$(PROJECT_NAME)/bin/$(cmd)
	sed -i "s/VERSION/$(VERSION)/g" "$(1)_pkg_build/debian/changelog"
	sed -i "s/DATE/$(DATE)/g" "$(1)_pkg_build/debian/changelog"
	@echo "================================="
	@echo ">>> Building package"
	cd $(1)_pkg_build && dpkg-buildpackage -d -rfakeroot
	@echo "================================="
endef
$(foreach cmd, $(CMDS), $(eval $(call PACKAGE_template,$(cmd))))

define RUN_template =
run-$(1): CMD_NAME=$(1)
run-$(1): CGO_ENABLED=$(CGO_ENABLED)
run-$(1): proto
	GOOS=$(OS) GOARCH=$(ARCH) $(CC) run -race $$(GO_CC_FLAGS) $$(CMD_PKG_PATH) $$(ARGS)
endef
$(foreach cmd, $(CMDS), $(eval $(call RUN_template,$(cmd))))

define DOCKERFILE_template =
.PHONY: dockerfile-$(1)
dockerfile-$(1): $(foreach cmd,$(CMDS),dockerfile-$(cmd)-$(1))
endef
$(foreach env,$(ENVS),$(eval $(call DOCKERFILE_template,$(env))))

# 1-cmd 2-env
define DOCKERFILE_CMD_template =
.PHONY: dockerfile-$(1)-$(2)
dockerfile-$(1)-$(2): FILE=Dockerfile.$(2).in
dockerfile-$(1)-$(2): CMD_NAME=$(1)
dockerfile-$(1)-$(2):
	$$(call check_defined, $$(VARS))
	sed -E $$(PARAMS) $$(FILE) > Dockerfile.$(1)
endef
$(foreach env,$(ENVS),$(foreach cmd,$(CMDS),$(eval $(call DOCKERFILE_CMD_template,$(cmd),$(env)))))

define DOCKER_BUILD_template =
.PHONY: docker-build-$(1)
docker-build-$(1): $(foreach cmd,$(CMDS),docker-build-$(cmd)-$(1))
endef
$(foreach env,$(ENVS),$(eval $(call DOCKER_BUILD_template,$(env))))

# 1-cmd 2-env
define DOCKER_BUILD_CMD_template =
.PHONY: docker-build-$(1)-$(2)
docker-build-$(1)-$(2): FILE=Dockerfile.$(2).in
docker-build-$(1)-$(2): CMD_NAME=$(1)
docker-build-$(1)-$(2): dockerfile-$(1)-$(2)
	docker build -t $(REGISTRY)/$(1):$(VERSION) -f Dockerfile.$(1) .
endef
$(foreach env,$(ENVS),$(foreach cmd,$(CMDS),$(eval $(call DOCKER_BUILD_CMD_template,$(cmd),$(env)))))

tools-install:
	cat tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -tI % go install %

cmds:
	@echo $(CMDS)

define ECHO_template =
.PHONY: $(1)
$(1): CMD_NAME=$(2)
$(1):
	@echo $(3)
endef
$(foreach cmd, $(CMDS), $(eval $(call ECHO_template,compose-image-tag-$(cmd),$(cmd),$$(CMD_IMAGE))))
$(foreach cmd, $(CMDS), $(eval $(call ECHO_template,compose-go-project-path-$(cmd),$(cmd),$$(GO_PROJECT_PACKAGE))))

$(BIN_DIR):
	mkdir -p $(BIN_DIR)

.PHONY: test
test: TEST_RUN?=^.*$$
test: TEST_DIR?=./...
test:
	$(CC) test \
		-race \
        -timeout 1h \
		-coverprofile cp.out \
		-run '$(TEST_RUN)' \
		$(TEST_DIR)

.PHONY: test-coverage
test-coverage: test
	@echo "================================================"
	@echo "=================TOTAL COVERAGE================="
	@echo "================================================"
	@$(CC) tool cover -func cp.out | grep total | awk '{print "coverage: " $$3 " of statements"}'

.PHONY: lint
lint:
	@golangci-lint run --timeout 5m -D structcheck,unused -E bodyclose,exhaustive,exportloopref,gosec,misspell,rowserrcheck,unconvert,unparam --out-format tab --sort-results --tests=false

.PHONY: docs
docs:
	swag init  --parseDependency -d ./cmd/snacker -o ./docs/rest-api/snacker/ --parseDepth 3

.PHONY: run-docs
run-docs: docs
	python -m http.server --directory docs/

# sedder for sonar-properties
# generates sonar-properties target that seds ./sonar-project.properties.in with version etc. with VERSION etc -> output to ./sonar-project.properties
$(eval $(call SEDDER_TARGET_template,sonar-properties,sonar-project.properties.in,,sonar-project.properties))

clean:
	docker-compose down
	rm -rf $(BIN_DIR)/*
	rm -rf $(foreach cmd,$(CMDS), $(cmd)_pkg_build)
	rm -rf $(shell find proto -type f -name '*.pb.go') \
		$(shell find proto -type f -name '*.pb.gw.go') \
		$(shell find proto -type f -name '*.swagger.json') \
	rm -rf Dockerfile $(foreach cmd,$(CMDS), Dockerfile.$(cmd))
	rm -rf sonar-project.properties cp.out
	rm -rf $(foreach cmd,$(CMDS), ./cmd/$(cmd)/pkged.go)
	rm -rf ./docs/rest-api/snacker/*

stats:
	scc --exclude-dir 'vendor,node_modules,data,.git,docker/etcdkeeper,utils' --wide
