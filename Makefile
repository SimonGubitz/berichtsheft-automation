
build:
	mkdir -p bin
	go build -o ./bin/berichtsheft src/main.go

check-format:
	test -z "$$(gofmt -l .)"

lint:
	golangci-lint run

test:
	go test -count=1 -race -shuffle=on ./...

vet:
	go vet ./...

ci: check-format lint test vet

