.Phony: init
init:
	test -f .env || cp .env.template .env

.PHONY: lint
lint:
	golangci-lint run --config ".github/.golangci.yml" --fast ./...
