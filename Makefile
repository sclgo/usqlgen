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
lint:
	golangci-lint run -v

checks: check_tidy check_vuln check_modern

check_tidy:
	go mod tidy
	# Verify that `go mod tidy` didn't introduce any changes. Run go mod tidy before pushing.
	git diff --exit-code --stat go.mod go.sum

.PHONY: check_vuln
check_vuln:
	go run golang.org/x/vuln/cmd/govulncheck@v1.1.4 ./...

.PHONY: check_modern
check_modern:
	go run golang.org/x/tools/gopls/internal/analysis/modernize/cmd/modernize@v0.20.0 ./...
# non-zero exit status on issues found
# nb: modernize is not part of golangci-lint yet - https://github.com/golangci/golangci-lint/issues/686
