# Strip debug info
GO_FLAGS += "-ldflags=-s -w"

# Avoid embedding the build path in the executable for more reproducible builds
GO_FLAGS += -trimpath

.PHONY: list build vet fmt test update-deps optimize install

list: #list all commands
	@echo "Commands:" && grep '^[^#[:space:]].*:' Makefile | cut -d'.' -f1 | awk NF | cut -d':' -f1

build: cmd/nrb/*.go lib/*/*.go go.mod #build cli
	CGO_ENABLED=0 go build $(GO_FLAGS) ./cmd/nrb

vet: #go vet project
	go vet ./cmd/... ./lib/...

fmt: #format project
	go fmt ./cmd/... ./lib/...

test: #do "testing"
	@echo "no testing"

update-deps: #update project deps
	go get -u ./...

optimize: #strip binary
	@command -v upx >/dev/null 2>&1 && upx ./nrb* || echo "install 'upx' to PATH"

install: #install project as cli bin
	go install ./...
