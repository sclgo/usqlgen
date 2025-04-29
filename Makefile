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

#NB: CI uses the golangci-lint Github action, not this target
.PHONY: lint
lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.5 run -v

.PHONY: check_vuln
check_vuln:
	go run golang.org/x/vuln/cmd/govulncheck@v1.1.4 ./...
# if we use more tools, we can switch to go tool -modfile=tools.mod
# there is good discussion at https://news.ycombinator.com/item?id=42845323

check_tidy:
	go mod tidy
	# Verify that `go mod tidy` didn't introduce any changes. Run go mod tidy before pushing.
	git diff --exit-code --stat go.mod go.sum