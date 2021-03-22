.Phony: init
init:
	test -f .env || cp .env.template .env
