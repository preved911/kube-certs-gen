BINARY := kube-certs-gen

.DEFAULT_GOAL := release

.PHONY: windows
windows:
	mkdir -p release
	GOOS=windows GOARCH=amd64 go build -o release/$(BINARY)-${TRAVIS_TAG}-windows-amd64

.PHONY: linux
linux:
	mkdir -p release
	GOOS=linux GOARCH=amd64 go build -o release/$(BINARY)-${TRAVIS_TAG}-linux-amd64

.PHONY: darwin
darwin:
	mkdir -p release
	GOOS=darwin GOARCH=amd64 go build -o release/$(BINARY)-${TRAVIS_TAG}-darwin-amd64

.PHONY: release
release: windows linux darwin
