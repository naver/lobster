VERSION = v1.0.0
REGISTRY ?= 
BASE_IMAGE ?= alpine

TAG_LOBSTER_STORE ?= $(REGISTRY)/lobster-store:$(VERSION)
DOCKERFILE_LOBSTER_STORE ?= build/lobster-store/Dockerfile

TAG_LOBSTER_QUERY ?= $(REGISTRY)/lobster-query:$(VERSION)
DOCKERFILE_LOBSTER_QUERY ?= build/lobster-query/Dockerfile

TAG_LOBSTER_GLOBAL_QUERY ?= $(REGISTRY)/lobster-global-query:$(VERSION)
DOCKERFILE_LOBSTER_GLOBAL_QUERY ?= build/lobster-global/Dockerfile

TAG_LOBSTER_SYNCER ?= $(REGISTRY)/lobster-syncer:$(VERSION)
DOCKERFILE_LOBSTER_SYNCER ?= build/lobster-syncer/Dockerfile

TAG_LOBSTER_EXPORTER ?= $(REGISTRY)/lobster-exporter:$(VERSION)
DOCKERFILE_LOBSTER_EXPORTER ?= build/lobster-exporter/Dockerfile

TAG_LOBSTER_LOGGEN ?= $(REGISTRY)/loggen:$(VERSION)
DOCKERFILE_LOBSTER_LOGGEN ?= build/loggen/Dockerfile

TAG_LOBSTER_OPERATOR ?= $(REGISTRY)/lobster-operator:$(VERSION)
DOCKERFILE_LOBSTER_OPERATOR ?= build/lobster-operator/Dockerfile

##@ Development

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: lint
lint: golangci-linter
	$(GOLANGCI_LINTER) run --verbose

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./pkg/operator/..." output:dir=deploy/templates/operator/manifests

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="./pkg/operator/hack/boilerplate.go.txt" paths="./pkg/operator/..." 

.PHONY: postProcessManifests
postProcessManifests: 
	@echo "Refine manifests file with helm charts condition"
	@echo '{{- if .Values.operator }}' > deploy/templates/operator/manifests/crd.yaml
	@echo '{{- if .Values.operator }}' > deploy/templates/operator/manifests/clusterRole.yaml
	@cat deploy/templates/operator/manifests/*lobstersinks.yaml >> deploy/templates/operator/manifests/crd.yaml
	@cat deploy/templates/operator/manifests/role.yaml >> deploy/templates/operator/manifests/clusterRole.yaml
	@echo '{{- end }}' >> deploy/templates/operator/manifests/crd.yaml
	@echo '{{- end }}' >> deploy/templates/operator/manifests/clusterRole.yaml
	@rm deploy/templates/operator/manifests/*lobstersinks.yaml
	@rm deploy/templates/operator/manifests/role.yaml
	

##@ Build Operator Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
GOLANGCI_LINTER ?= $(LOCALBIN)/golangci-lint
SWAG ?= $(LOCALBIN)/swag
WIDDERSHINS ?= widdershins

## Tool Versions
CONTROLLER_TOOLS_VERSION ?= v0.14.0
GOLANGCI_LINT_TOOLS_VERSION ?= v1.59.1
SWAG_VERSION := v1.16.3
WIDDERSHINS_TOOLS_VERSION ?= 4.0.1

OS := $(shell uname | tr '[:upper:]' '[:lower:]')
ARCH := $(shell uname -m)

check_command = $(shell which $1 > /dev/null 2>&1 && echo "found" || echo "not found")

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	@GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: golangci-linter
golangci-linter: $(GOLANGCI_LINTER) ## Download golangci-linter locally if necessary.
$(GOLANGCI_LINTER): $(LOCALBIN)
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(LOCALBIN) $(GOLANGCI_LINT_TOOLS_VERSION)

.PHONY: swag
swag: ## Install swag if necessary.
$(SWAG): $(LOCALBIN)
	@GOBIN=$(LOCALBIN) go install github.com/swaggo/swag/cmd/swag@$(SWAG_VERSION)

.PHONY: widdershins
widdershins: ## Install widdershins if necessary.
ifeq ($(call check_command,$(WIDDERSHINS)),not found)
	@echo "$(WIDDERSHINS) is not found, installing via npm..."
	@npm install -g widdershins@$(WIDDERSHINS_TOOLS_VERSION)
endif

##@ Generate docs

.PHONY: gen-swagger-docs-query
gen-swagger-docs-query: fmt vet swag
	$(SWAG) fmt -d cmd/lobster-query,pkg/lobster/server/handler/log
	$(SWAG) init -d cmd/lobster-query,pkg/lobster/server/handler/log -g main.go -o pkg/docs/query/ --parseDependency true --tags Post
	cp pkg/docs/query/swagger.json web/static/docs/query/

.PHONY: gen-swagger-docs-global-query
gen-swagger-docs-global-query: fmt vet swag
	$(SWAG) fmt -d cmd/lobster-global,pkg/lobster/server/handler/log
	$(SWAG) init -d cmd/lobster-global,pkg/lobster/server/handler/log -g main.go -o pkg/docs/global-query/ --parseDependency true --tags Post
	cp pkg/docs/global-query/swagger.json web/static/docs/global-query/

.PHONY: gen-swagger-docs-operator
gen-swagger-docs-operator: manifests postProcessManifests generate fmt vet swag
	$(SWAG) fmt -d pkg/operator/server,pkg/operator/server/handler
	$(SWAG) init -d pkg/operator/server,pkg/operator/server/handler -g server.go -o pkg/docs/operator/ --parseDependency true --tags Get,Put,Delete
	cp pkg/docs/operator/swagger.json web/static/docs/operator/

.PHONY: gen-swagger-docs
gen-swagger-docs: gen-swagger-docs-query gen-swagger-docs-global-query gen-swagger-docs-operator

.PHONY: gen-swagger-docs-query
gen-api-docs-query: gen-swagger-docs-query widdershins
	$(WIDDERSHINS) ./pkg/docs/query/swagger.yaml --code=true --shallowSchemas=true --omitHeader=true --summary=true --lang=true  -o ./docs/apis/query_apis.md

.PHONY: gen-api-docs-global-query
gen-api-docs-global-query: gen-swagger-docs-global-query widdershins
	$(WIDDERSHINS) ./pkg/docs/global-query/swagger.yaml --code=true --shallowSchemas=true --omitHeader=true --summary=true --lang=true  -o ./docs/apis/global_query_apis.md

.PHONY: gen-api-docs-operator
gen-api-docs-operator: gen-swagger-docs-operator widdershins
	$(WIDDERSHINS) ./pkg/docs/operator/swagger.yaml --code=true --shallowSchemas=true --omitHeader=true --summary=true --lang=true  -o ./docs/apis/operator_apis.md

.PHONY: gen-api-docs
gen-api-docs: gen-api-docs-query gen-api-docs-global-query gen-api-docs-operator

##@ Build

.PHONY: image-store image-query image-global-query image-loggen image-exporter
image-store: fmt vet
	docker build  --build-arg BASE_IMAGE=${BASE_IMAGE} \
	-t $(TAG_LOBSTER_STORE) -f $(DOCKERFILE_LOBSTER_STORE) . 

image-query: gen-swagger-docs-query
	docker build --build-arg BASE_IMAGE=${BASE_IMAGE}  \
	-t $(TAG_LOBSTER_QUERY) -f $(DOCKERFILE_LOBSTER_QUERY) . 

image-global-query: gen-swagger-docs-global-query
	docker build --build-arg BASE_IMAGE=${BASE_IMAGE}  \
	-t $(TAG_LOBSTER_GLOBAL_QUERY) -f $(DOCKERFILE_LOBSTER_GLOBAL_QUERY) . 

image-syncer: fmt vet
	docker build --build-arg BASE_IMAGE=${BASE_IMAGE}  \
	-t $(TAG_LOBSTER_SYNCER) -f $(DOCKERFILE_LOBSTER_SYNCER) . 

image-loggen: fmt vet
	docker build --build-arg BASE_IMAGE=${BASE_IMAGE}  \
	-t $(TAG_LOBSTER_LOGGEN) -f $(DOCKERFILE_LOBSTER_LOGGEN) . 

image-exporter: fmt vet
	docker build --build-arg BASE_IMAGE=${BASE_IMAGE}  \
	-t $(TAG_LOBSTER_EXPORTER) -f $(DOCKERFILE_LOBSTER_EXPORTER) . 

image-operator: manifests generate fmt vet gen-swagger-docs-operator
	docker build --build-arg BASE_IMAGE=${BASE_IMAGE}  \
	-t ${TAG_LOBSTER_OPERATOR} -f $(DOCKERFILE_LOBSTER_OPERATOR) . 

.PHONY: push-store push-query push-global-query push-loggen push-exporter push-operator
push-store: image-store
	docker push $(TAG_LOBSTER_STORE)
	
push-query: image-query
	docker push $(TAG_LOBSTER_QUERY)

push-global-query: image-global-query
	docker push $(TAG_LOBSTER_GLOBAL_QUERY)

push-syncer: image-syncer
	docker push $(TAG_LOBSTER_SYNCER)

push-loggen: image-loggen
	docker push $(TAG_LOBSTER_LOGGEN)

push-exporter: image-exporter
	docker push $(TAG_LOBSTER_EXPORTER)

push-operator: image-operator
	docker push $(TAG_LOBSTER_OPERATOR)

.PHONY: push-all
push-all: push-store push-query push-global-query push-syncer push-exporter push-loggen push-operator

.PHONY: version
version:
	@echo $(VERSION)

.PHONY: tag
tag:
	git tag $(VERSION)
	git push origin $(VERSION)
