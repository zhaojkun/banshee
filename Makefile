export GO15VENDOREXPERIMENT=1
GO_PKGS=$(shell GO15VENDOREXPERIMENT=1 go list ./... | grep -v '/vendor/')

default: build

build:
	go build

deps:
	godep get -t ${GO_PKGS}
	godep save -t ${GO_PKGS}

test: lint vet
	go test $(GO_PKGS)

lint:
	for line in $(GO_PKGS); do fgt golint "$$line" || exit 1; done

vet:
	go vet $(GO_PKGS)

changelog:
	git log --first-parent --pretty="format:* %b" v`./banshee -v`..

static:
	make -C static deps
	make -C static build

docker:
	cd docker; docker build --no-cache -t banshee .

.PHONY: deps test lint vet changelog static docker
