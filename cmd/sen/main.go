package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/briheet/sen/internal/cmd"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	ret := cmd.Execute(ctx)
	os.Exit(ret)
}
