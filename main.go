package main

import "context"

var (
	version = "development"
)

func main() {

	ctx := context.Background()

	Execute(ctx, version)

}
