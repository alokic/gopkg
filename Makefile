# Set an output prefix, which is the local directory if not specified
PREFIX?=$(shell pwd)

GITHUB_USERNAME=alokic
APPNAME=gopkg
PROJECT_ROOT=${GOPATH}/src/github.com/${GITHUB_USERNAME}/${APPNAME}
SCRIPT_FOLDER=${PROJECT_ROOT}/scripts
GOBIN=${GOPATH}/bin

.PHONY: all dep fmt vet build test cover tag help checkversion
all: dep fmt vet build test

dep: 	## Get all deps
	@echo "Running $@"
	@dep ensure -v

test: 	## Tests the project except vendor and deployment folders
	@echo "Running $@"
	@go test $(shell go list ./... | grep -v /vendor/ | grep -v /deployment/ | grep -v /output/ )

lint:														## lints the project except vendor and deployment folders
	@echo "Running $@"
	@golint $(shell go list ./... | grep -v /vendor/ | grep -v /deployment/ |  grep -v /output/) | grep -v '.pb.go:' | tee /dev/stderr


vet:														## Vets the project except vendor and deployment folders
	@echo "Running $@"
	@go vet $(shell go list ./... | grep -v /vendor/ | grep -v /deployment/ |  grep -v /output/) | grep -v '.pb.go:' | tee /dev/stderr


fmt:														## Formats the project except vendor and deployment folders
	@echo "Running $@"
	@go fmt  $(shell go list ./... | grep -v /vendor/ | grep -v /deployment/ |  grep -v /output/ | grep -v '.pb.go:')


cover: ## Runs go test with coverage
	@echo "Running $@"
	@echo "" > coverage.txt
	@for d in $(shell go list ./... | grep -v /vendor/ | grep -v /deployment/ |  grep -v /output/ | grep -v '.pb.go:'); do \
		go test -race -coverprofile=profile.out -covermode=atomic "$$d"; \
		if [ -f profile.out ]; then \
			cat profile.out >> coverage.txt; \
			rm profile.out; \
		fi; \
	done;

clean:														## Clean any stray files formed during make
	@echo "Running $@"


tag: checkversion ## Create a new git tag to prepare to build a release
	@echo "Running $@"
	git tag -sa $(VERSION) -m "$(VERSION)"
	@echo "Run git push origin $(VERSION) to push your new tag to GitHub and trigger a travis build."

help:  ## Print help
	@echo "=================================================="
	@echo "Run: make <target_name> NAMESPACE=<namespace_name>"
	@echo "=================================================="
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

checkversion:
ifeq ($(VERSION),)
	@echo "Missing VERSION"
	@exit 1
endif


build:
	@echo "Running $@"
	@go build ./...