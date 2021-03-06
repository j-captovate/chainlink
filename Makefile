.DEFAULT_GOAL := build
.PHONY: dep build install

LDFLAGS=-ldflags "-X github.com/smartcontractkit/chainlink/store.Sha=`git rev-parse HEAD`"

dep:
	@dep ensure

build: dep
	@go build $(LDFLAGS) -o chainlink

install: dep
	@go install $(LDFLAGS)
