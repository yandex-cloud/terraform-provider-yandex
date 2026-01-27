SEMVER ?= 0.0.1

TEST?=$$(go list ./... )
GOFMT_FILES?=$$(find . -name '*.go')
PKG_NAME=yandex
LINT_PACKAGES= ./yandex/... yandex-framework/...

SWEEP?=$(YC_REGION)
ifeq ($(SWEEP),)
SWEEP=ru-central1
endif

SWEEP_DIR= ./yandex ./yandex-framework/...

SWEEPERS_FOR_RUNNING?=""

VCS_TYPE := $(shell arc info 2>&1 | grep -q "Not a mounted arc repository" && echo "git" || echo "arc")
ifeq ($(VCS_TYPE),arc)
    # TODO: get real tag instead of revision number
    version_tag := $(shell arc describe --svn | cut -d'-' -f1 | cut -d'r' -f2 | tr -d ' \n')
    commit_hash := $(shell arc rev-parse HEAD | head -c 8)
else
    version_tag := $(shell git describe --abbrev=0 --tags)
    commit_hash := $(shell git rev-parse --short HEAD)
endif

current_time = $(shell date +"%Y-%m-%dT%H-%M-%SZ")
LDFLAGS = -ldflags "-s -w -X github.com/yandex-cloud/terraform-provider-yandex/version.ProviderVersion=${version_tag}-${current_time}+dev.${commit_hash}"


TFGEN_MK := ./tools/tfgen/gen.mk
-include $(TFGEN_MK)

# Define fallback targets if the file doesn't exist
ifeq ($(wildcard $(TFGEN_MK)),)
    # Define dummy targets or alternative behavior
    generate-public:
	    @echo "tfgen.mk not found, skipping tfgen operations"
endif

default: build

##build and local-build generate the same user-agent
##The difference is that build will put the file in the $GOPATH/bin folder, while local-build will put it in the tf plugins folder
##Example user-agent: Terraform/1.5.7 (https://www.terraform.io) terraform-provider-yandex/0.124.0-2024-07-26T16-24-43Z+dev.79152e6a grpc-go/1.62.1
build: fmtcheck
	go install ${LDFLAGS}

local-build: fmtcheck
	go build ${LDFLAGS} -o $(HOME)/.terraform.d/plugins/registry.terraform.io/yandex-cloud/yandex/$(SEMVER)/$(shell go env GOOS)_$(shell go env GOARCH)/terraform-provider-yandex main.go

sweep:
	@echo "WARNING: This will destroy infrastructure. Use only in development accounts.";
	go test $(SWEEP_DIR) -v -sweep=$(SWEEP) -sweep-run=$(SWEEPERS_FOR_RUNNING) -timeout 60m

test: fmtcheck
	go test $(TEST) -timeout=60s -parallel=4

testacc: fmtcheck
	TF_ACC=1 TF_SCHEMA_PANIC_ON_ERROR=1 go test $(TEST) -v $(TESTARGS) -timeout 120m

vet:
	@echo "go vet ."
	@go vet $$(go list ./...) ; if [ $$? -eq 1 ]; then \
		echo ""; \
		echo "Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for review."; \
		exit 1; \
	fi

fmt:
	gofmt -w $(GOFMT_FILES)

fmtcheck:
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

lint:
	golangci-lint version
	golangci-lint run --modules-download-mode mod $(LINT_PACKAGES) -v

tools:
	@echo "==> installing required tooling..."
	go install github.com/client9/misspell/cmd/misspell
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.2.2
	go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

test-compile:
	@if [ "$(TEST)" = "./..." ]; then \
		echo "ERROR: Set TEST to a specific package. For example,"; \
		echo "  make test-compile TEST=./$(PKG_NAME)"; \
		exit 1; \
	fi
	go test -c $(TEST) $(TESTARGS)


changie-lint:
	go run lint/cmd/changie/changie.go batch patch -d

install-yfm:
	npm i @diplodoc/cli -g

generate-docs:
	rm -rf templates-[0-9]*
	go run tools/cmd/generate-docs/generate_docs.go ./templates ./docs

affected-lint-provider-docs:
	@sh -c "'$(CURDIR)//scripts/affectedocs.sh'"

build-website: generate-docs
	go run tools/cmd/generate-toc/generate_toc.go ./docs && \
 	yfm -i ./docs -o ./output-folder -c .yfm -v '{"version": "$(SEMVER)"}'

# to run this command please set YFM_STORAGE_SECRET_KEY and YFM_STORAGE_KEY_ID of the bucket
publish-website: generate-docs
	go run tools/cmd/generate-toc/generate_toc.go ./docs && \
	yfm -i ./docs -o ./output-folder -c .yfm -v '{"version": "$(SEMVER)"}' --publish

validate-docs:
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs validate -provider-name ${PKG_NAME}

generate: generate-public-api-desc generate-public generate-docs

.PHONY: build sweep test testacc vet fmt fmtcheck lint tools test-compile website changie-lint build-website publish-website generate-docs install-yfm affected-lint-provider-docs generate
