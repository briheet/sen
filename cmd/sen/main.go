package main

import (
	"context"
	"os"
	"os/signal"

	rustbuild "github.com/briheet/sen/internal/adapters/rust/build"
	"github.com/briheet/sen/internal/cmd"
)

func main() {
	if os.Getenv(rustbuild.WrapperEnv) == "1" {
		os.Exit(rustbuild.RunWrapper(os.Args[1:]))
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	ret := cmd.Execute(ctx)
	os.Exit(ret)
}
