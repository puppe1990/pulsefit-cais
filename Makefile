.PHONY: dev build test css css-watch install console

CAIS := $(shell command -v cais 2>/dev/null || command -v $(HOME)/go/bin/cais 2>/dev/null)

BIN := bin/server
CSS_IN := input.css
CSS_OUT := web/static/css/styles.css

install:
	$(CAIS) install

console:
	$(CAIS) console

test:
	go test ./... -race -count=1

css:
	npx tailwindcss -i $(CSS_IN) -o $(CSS_OUT) --minify

css-watch:
	npx tailwindcss -i $(CSS_IN) -o $(CSS_OUT) --watch

build: css
	CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BIN) ./cmd/server

dev: css
	$(MAKE) css-watch &
	$(CAIS) dev

# cais quality tooling
.PHONY: lint format format-check pre-commit-install ci

lint:
	golangci-lint run ./...

format:
	npm run format

format-check:
	npm run format:check

pre-commit-install:
	pre-commit install

ci: test lint format-check