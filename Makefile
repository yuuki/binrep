COMMIT = $$(git describe --always)
PKG = github.com/yuuki/binrep
PKGS = $$(go list ./... | grep -v vendor)

all: build

.PHONY: build
build: deps generate
	go build -ldflags "-X main.GitCommit=\"$(COMMIT)\"" $(PKG)

.PHONY: test
test: vet
	go test -v $(PKGS)

.PHONY: vet
vet:
	go vet $(PKGS)

.PHONY: lint
lint:
	golint $(PKGS)

.PHONY: deps
deps:
	go get github.com/jteeuwen/go-bindata/...

.PHONY: generate
generate:
	touch CREDITS
	go generate -x ./...

.PHONY: credits
credits:
	scripts/credits > CREDITS

.PHONY: release
release: credits generate
	scripts/release
