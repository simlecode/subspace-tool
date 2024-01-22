git=$(subst -,.,$(shell git describe --always --match=NeVeRmAtCh --dirty 2>/dev/null || git rev-parse --short HEAD 2>/dev/null))

ldflags=-X=github.com/simlecode/subspace-tool/version.CurrentCommit=+git.$(git)
ifneq ($(strip $(LDFLAGS)),)
	ldflags+=-extldflags=$(LDFLAGS)
endif

GOFLAGS+=-ldflags="$(ldflags)"

all: collect block-collect
.PHONY: all

collect:
	rm -rf ./collect
	go build $(GOFLAGS) -o collect ./cmd/collect
.PHONY: collect

block-collect:
	rm -rf ./block-collect
	go build $(GOFLAGS) -o block-collect ./cmd/blockCollect
.PHONY: block-collect

lint:
	golangci-lint run

test:
	go test -race ./...