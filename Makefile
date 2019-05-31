WD := $(shell pwd)
GOPATH := $(shell go env GOPATH)

export GO111MODULE=on

TEST_FLAGS := ""

default: test

getdeps:
	@echo "Installing golint" && go get -u golang.org/x/lint/golint
	@echo "Installing gometalinter" && go get -u github.com/alecthomas/gometalinter

verifiers: lint metalinter

fmt:
	@echo "Running $@"
	@${GOPATH}/bin/golint .

lint:
	@echo "Running $@"
	@${GOPATH}/bin/golint -set_exit_status ./...

metalinter:
	@${GOPATH}/bin/gometalinter --install
	@${GOPATH}/bin/gometalinter --disable-all \
		-E vet \
		-E gofmt \
		-E misspell \
		-E ineffassign \
		-E goimports \
		-E deadcode --tests --vendor ./...

check: verifiers test

test:
	@echo "Running unit tests"
	@go test -v $(TEST_FLAGS) -tags kqueue ./...

bench:
	@echo "Running bench"
	@go test -bench=. -benchmem -benchtime=5s ./...

coverage:
	@curl -s https://codecov.io/bash | bash

clean:
	@echo "Cleaning up all the generated files"
	@find . -name '*.test' | xargs rm -fv
	@rm coverage.txt
	@rm -rvf build
	@rm -rvf release