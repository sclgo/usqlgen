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


