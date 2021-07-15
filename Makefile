WD := $(shell pwd)
GOPATH := $(shell go env GOPATH)

export GO111MODULE=on

TEST_FLAGS := ""

default: fmt lint test

getdeps:
	@echo "Installing golint" && go get -u golang.org/x/lint/golint

verifiers: lint

fmt:
	@echo "Running $@"
	@${GOPATH}/bin/golint .

lint:
	@echo "Running $@"
	@${GOPATH}/bin/golint -set_exit_status ./...

check: verifiers test

test:
	@echo "Running unit tests"
	@go test -v $(TEST_FLAGS) -tags kqueue ./...

bench:
	@echo "Running bench"
	@go test -bench=. -benchmem -benchtime=5s ./...

coverage:
	@bash ./gen-coverage.sh

clean:
	@echo "Cleaning up all the generated files"
	@find . -name '*.test' | xargs rm -fv
	@rm coverage.txt
	@rm -rvf build
	@rm -rvf release