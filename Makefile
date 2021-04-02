GOLANGCI_LINT_VERSION := 1.39.0

.Phony: init
init:
	test -f .env || cp .env.template .env

.PHONY: lint
lint:
	./bin/golangci-lint run --config ".github/.golangci.yml" --fast ./...

.PHONY: bootstrap_golangci_lint
bootstrap_golangci_lint:
	mkdir -p bin
	curl -L -o ./bin/golangci-lint.tar.gz https://github.com/golangci/golangci-lint/releases/download/v$(GOLANGCI_LINT_VERSION)/golangci-lint-$(GOLANGCI_LINT_VERSION)-$(shell uname -s)-amd64.tar.gz
	cd ./bin && \
	tar xzf golangci-lint.tar.gz && \
	mv golangci-lint-$(GOLANGCI_LINT_VERSION)-$(shell uname -s)-amd64/golangci-lint golangci-lint && \
	rm -rf golangci-lint-$(GOLANGCI_LINT_VERSION)-$(shell uname -s)-amd64 *.tar.gz
