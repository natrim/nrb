
# Set lib.Version variable with current hash
PACKAGE := github.com/natrim/nrb
VERSION := $(shell git describe --abbrev=0 --tags)
LDFLAGS := -X '$(PACKAGE)/lib.Version=$(VERSION)'

# Strip debug info (-s -w)
GO_FLAGS += -ldflags="$(LDFLAGS) -s -w"

# Avoid embedding the build path in the executable for more reproducible builds
GO_FLAGS += -trimpath

.PHONY: list build fmt format test update-deps install lint tidy analyse

list: #list all commands
	@echo "Commands:" && grep '^[^#[:space:]].*:' Makefile | cut -d'.' -f1 | awk NF | cut -d':' -f1

build: cmd/nrb/*.go lib/*/*.go go.mod #build cli
	CGO_ENABLED=0 go build $(GO_FLAGS) ./cmd/nrb

analyse: #escape analysis
	CGO_ENABLED=0 go run -ldflags="$(LDFLAGS)" -gcflags="-m" -trimpath ./cmd/nrb -v

lint: #lint project
	staticcheck ./...
	go vet ./cmd/... ./lib/...

format: #format project
	go fmt ./cmd/... ./lib/...
fmt: format #format project alias

test: #do "testing"
	@echo "🚫 no testing"

tidy: #tidy up go modules
	go mod tidy
	@echo "✔︎ Tidy complete"

update-deps: #update project deps
	go get -u ./...

install: tidy #install project as cli bin
	go install $(GO_FLAGS) ./...

release: #release binary on github
	goreleaser release --clean
