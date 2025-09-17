SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c # -c: Needed in .SHELLFLAGS. Default is -c.
.DEFAULT_GOAL := build

dst := tw

all: clean tidy fmt lint build test
clean:
	@echo "==> Cleaning" >&2
	@rm -f $(dst)
	@go clean -cache -testcache
tidy:
	@echo "==> Running go mod tidy -v"
	@go mod tidy -v
tidy-go:
	@v=$(shell go version|awk '{print $$3}' |sed -e 's,go\(.*\)\..*,\1,g') && go mod tidy -go=$${v}
deps:
	@go list -m all
update:
	@go get -u ./...
fmt:
	@echo "==> Running go fmt ./..." >&2
	@go fmt ./...
lint:
	@echo "==> Running golangci-lint run" >&2
	@golangci-lint run
build:
	@echo "==> Go Building" >&2
	@env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o $(dst) .
run:
	@cat <./test.txt | go run .

pkg := ./...
cover_mode := atomic
cover_out := cover.out
test: testsum-cover-check
testsum:
	@echo "==> Running go testsum" >&2
	@gotestsum --format testname -- -v $(pkg) -coverprofile=$(cover_out) -covermode=$(cover_mode) -coverpkg=$(pkg)
testsum-cover-check: testsum
	@echo "==> Running test-coverage" >&2
	@go-test-coverage --config=./.testcoverage.yaml

tag:
	@v=$$(git tag --list |sort -V |tail -1) && nv="$${v%.*}.$$(($${v##*.}+1))" && echo "==> New tag: $${nv}" && git tag $${nv}
tagp: tag
	@git push --tags


gr_init:
	@goreleaser init
gr_check:
	@goreleaser check
gr_snap:
	@goreleaser release --snapshot --clean $(OPT)
gr_snap_skip_publish:
	@OPT=--skip-publish make gr_snap
gr_build:
	@goreleaser build --snapshot --clean
