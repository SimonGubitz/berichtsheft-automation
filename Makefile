TARGET := $(word 2, $(MAKECMDGOALS))
SRC := cmd/$(TARGET)/main.go

run:
	export $$(cat .env.local | xargs) && go run $(SRC)

build:
	go build -o ./bin/main-$(TARGET) $(SRC)


check-format:
	test -z "$$(gofmt -l .)"

lint:
	golangci-lint run

test:
	go test -count=1 -race -shuffle=on ./...

vet:
	go vet ./...

ci: check-format lint test vet

