VERSION = 0.0.6
SOURCE = ./...

.DEFAULT_GOAL := build

help:
	# build:     build terraputs (default make target)
	# tools: 		 install build dependencies
	# vet:       run 'go vet' against source code
	# test:      run automated tests
	# check-tag: check if a $(VERSION) git tag already exists
	# tag:       create a $(VERSION) git tag
	# release:   build and publish a terraputs GitHub release
	# clean:     remove compiled artifacts

tools:
	go install github.com/goreleaser/goreleaser@latest
.PHONY: tools

build: tools
	goreleaser release \
		--snapshot \
		--skip-publish \
		--rm-dist
.PHONY: build

vet:
	go vet $(SOURCE)
.PHONY: vet

test-fmt:
	test -z $(shell go fmt $(SOURCE))
.PHONY: test-fmt

test: vet test-fmt
	go test -cover $(SOURCE) -count=1
.PHONY: test

check-tag:
	./scripts/ensure-unique-version.sh "$(VERSION)"
.PHONY: check-tag

tag: check-tag
	@echo "creating git tag $(VERSION)"
	@git tag $(VERSION)
	@git push origin $(VERSION)
.PHONY: tag

release: tools
	goreleaser release \
		--rm-dist
.PHONY: release

demo:
	svg-term \
		--cast 423523 \
		--out demo.svg \
		--window \
		--no-cursor
.PHONY: demo

clean:
	rm -rf dist
.PHONY: clean
