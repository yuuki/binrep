COMMIT = $$(git describe --tags --always)
# date format of goreleaser
DATE = $$(date --utc '+%Y-%m-%d_%H:%M:%S')
PKG = github.com/yuuki/binrep
PKGS = $$(go list ./... | grep -v vendor)
CREDITS = vendor/CREDITS

all: build

.PHONY: build
build: deps
	go build -ldflags "-X main.commit=\"$(COMMIT)\" -X main.date=\"$(DATE)\"" $(PKG)

.PHONY: test
test: vet
	go test -v ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: lint
lint: devel-deps
	golint -set_exit_status $(PKGS)

.PHONY: deps
deps:
	go get github.com/jteeuwen/go-bindata/...

.PHONY: devel-deps
devel-deps: deps
	go get github.com/golang/lint/golint
	go get github.com/motemen/gobump
	go get github.com/Songmu/ghch

.PHONY: generate
generate:
	go generate -x ./...

.PHONY: credits
credits:
	scripts/credits > $(CREDITS)

.PHONY: release
release: credits generate
	scripts/release
