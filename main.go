package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/glennr/wt/cmd"
)

func main() {
	os.Exit(run())
}

func run() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := cmd.Execute(ctx); err != nil {
		return 1
	}
	return 0
}
