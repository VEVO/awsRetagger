# setting some defaults if those variables are empty
OWNER=vevo
APP_NAME=awsRetagger
IMAGE_NAME=$(OWNER)/$(APP_NAME)
GO_REVISION?=$(shell git rev-parse HEAD)
GO_TO_REVISION?=$(GO_REVISION)
GO_FROM_REVISION?=$(shell git rev-parse refs/remotes/origin/master)
GIT_TAG=$(IMAGE_NAME):$(GO_REVISION)
BUILD_VERSION?=$(shell date +%Y%m%d%H%M%S)-dev
BUILD_TAG=$(IMAGE_NAME):$(BUILD_VERSION)
LATEST_TAG=$(IMAGE_NAME):latest

PHONY: go-build

docker-lint:
	docker run -it --rm -v "${PWD}/Dockerfile":/Dockerfile:ro redcoolbeans/dockerlint

docker-login:
	@docker login -u "$(DOCKER_USER)" -p "$(DOCKER_PASS)"

go-dep:
	@if [ -f "glide.yaml" ] ; then \
		go get github.com/Masterminds/glide \
		&& go install github.com/Masterminds/glide \
		&& glide install --strip-vendor; \
	elif [ -f "Godeps/Godeps.json" ] ; then \
		go get github.com/tools/godep \
		&& godep restore; \
	else \
		go get -d -t -v ./...; \
	fi

GOFILES_NOVENDOR=$(shell find . -type f -name '*.go' -not -path "./vendor/*")

go-fmt:
	@[ $$(gofmt -l $(GOFILES_NOVENDOR) | wc -l) -gt 0 ] && echo "Code differs from gofmt's style" && exit 1 || true

go-lint: go-fmt
	@go get -u golang.org/x/lint/golint; \
	if [ -f "glide.yaml" ] ; then \
		golint -set_exit_status $$(glide novendor); \
		go vet -v $$(glide novendor); \
	else \
		golint -set_exit_status ./...; \
		go vet -v ./...; \
	fi

go-test:
	@if [ -f "glide.yaml" ] ; then \
		go test $$(glide novendor); \
	else \
		go test -v ./...; \
	fi

go-build: go-dep go-lint go-test
	@go build -v -a -ldflags "-X main.version=$(BUILD_VERSION)"

build: go-build

release:
	git tag -s $(BUILD_VERSION) -m "Release $(BUILD_VERSION)"
	# We skip publish for now for sanity purposes
	goreleaser release --rm-dist

# vim: ft=make
