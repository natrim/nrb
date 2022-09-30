# Strip debug info
GO_FLAGS += "-ldflags=-s -w"

# Avoid embedding the build path in the executable for more reproducible builds
GO_FLAGS += -trimpath

build: cmd/nrb/*.go lib/*/*.go go.mod
	CGO_ENABLED=0 go build $(GO_FLAGS) ./cmd/nrb

vet:
	go vet ./cmd/... ./lib/...

fmt:
	go fmt ./cmd/... ./lib/...

test:
	@echo "no testing"

update-deps:
	go get -u ./...

optimize:
	upx ./nrb*

install:
	go install ./...

