GOFILES_NOVENDOR=$(shell find . -type f -name '*.go' -not -path "./vendor/*")

default: dep lint test

dep:
	@if [ -f "glide.yaml" ] ; then \
		go get github.com/Masterminds/glide \
		&& go install github.com/Masterminds/glide \
		&& glide install --strip-vendor; \
	else \
		go get -v ./...; \
	fi

fmt:
	@[ $$(gofmt -l $(GOFILES_NOVENDOR) | wc -l) -gt 0 ] && echo "Code differs from gofmt's style" && exit 1 || true

lint: fmt
	@go get github.com/golang/lint/golint; \
	if [ -f "glide.yaml" ] ; then \
		golint -set_exit_status $$(glide novendor); \
		go vet -v $$(glide novendor); \
	else \
		golint -set_exit_status ./...; \
		go vet -v ./...; \
	fi

gocov:
	@go get github.com/axw/gocov/gocov \
	&& go install github.com/axw/gocov/gocov; \
	if [ -f "glide.yaml" ] ; then \
		gocov test $$(glide novendor) | gocov report; \
	else \
		gocov test ./... | gocov report; \
	fi
	# gocov test $$(glide novendor) >/tmp/gocovtest.json ; gocov annotate /tmp/gocovtest.json MyFunc

test:
	@if [ -f "glide.yaml" ] ; then \
		go test $$(glide novendor); \
	else \
		go test -v ./...; \
	fi

build: dep lint test
	go clean -v
	go build -v

install: dep
	go install
