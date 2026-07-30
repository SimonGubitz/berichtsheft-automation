SRC := ./cmd/$(PROJECT)/main.go

define HELP_BODY
berichtsheft
endef
export HELP_BODY

help:
	@echo "$$HELP_BODY"


run:
	set -a; . ./.env.local; set +a; time go run $(SRC)

build:
	mkdir -p bin
	go build -o ./bin/main-$(PROJECT) $(SRC)

debug:
	mkdir -p bin
	set -a; . ./.env.local; set +a; dlv debug $(SRC) --output ./bin/$(PROJECT)

check-format:
	test -z "$$(gofmt -l .)"

lint:
	golangci-lint run

test:
	go test -count=1 -race -shuffle=on ./...

vet:
	go vet ./...

ci: check-format lint test vet

