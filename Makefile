.PHONY: build
build:
	go build .

.PHONY: test
test: unit-test

.PHONY: unit-test
unit-test:
	go test -v -vet=all ./...

.PHONY: short-test
short-test:
	go test -v -short ./...

.PHONY: integration-test
integration-test:
	SUITE=integration go test -v -timeout 10m ./internal/integrationtest/$(SUBTEST)...

.PHONY: itest
itest: integration-test

.PHONY: lint
lint: tools/golangci-lint
	tools/golangci-lint run -v

.PHONY: checks
checks: check_tidy check_vuln check_modern

.PHONY: check_tidy
check_tidy:
	go mod tidy
	# Verify that `go mod tidy` didn't introduce any changes. Run go mod tidy before pushing.
	git diff --exit-code --stat go.mod go.sum

.PHONY: check_vuln
check_vuln:
	go run golang.org/x/vuln/cmd/govulncheck@v1.1.4 ./...

.PHONY: check_modern
check_modern:
	test -z "$$(go fix -diff ./... | tee /dev/stderr)"
# non-zero exit status on issues found
# nb: modernize is not part of golangci-lint yet - https://github.com/golangci/golangci-lint/issues/686

# Tools targets

tools:
	mkdir -p tools

tools/golangci-lint: tools
# Version must be the same as in golangci-lint Github action
# We install golangci-lint as recommended in the docs. See the same docs for a discussion about go run and
# go get -tool alternatives - https://golangci-lint.run/docs/welcome/install/ .
# Delete tools/golangci-lint if this target is updated (may be automated in the future)
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b ./tools v2.11.4