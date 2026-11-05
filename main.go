package main

import (
	"context"

	"berichtsheft/cmd"
)

var (
	version = "development"
)

func main() {

	ctx := context.Background()

	cmd.Execute(ctx, version)

}
