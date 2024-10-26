GO111MODULE = on
CGO_ENABLED = 0
GO_FILES = $(shell go list ./... | grep -v mocks)

.PHONY: test
test:
	go test -v $(GO_FILES) -coverprofile=coverage.out

.PHONY: test-s
test-s:
	go test -v $(pkg)

.PHONY: gen-mocks
gen-mocks:
	go generate $(GO_FILES)

.PHONY: cleantestcache
cleantestcache:
	go clean -testcache

.PHONY: tidy
tidy:
	GO111MODULE=$(GO111MODULE) go mod tidy

.PHONY: cover
cover: cleantestcache test
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func coverage.out

.PHONY: cleanlintcache
cleanlintcache:
	golangci-lint cache clean

.PHONY: lint
lint: cleanlintcache
	golangci-lint run --timeout 20m ./...

.PHONY: dev-tools
dev-tools:
	go install go.uber.org/mock/mockgen@v0.5.0
	./scripts/install_goreleaser.sh

.PHONY: test-all
test-all: cover lint

.PHONY: dist-dev
dist-dev:
	goreleaser build --snapshot --clean --single-target

