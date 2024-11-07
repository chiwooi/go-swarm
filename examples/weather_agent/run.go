package main

import(
    "context"

    "github.com/chiwooi/go-swarm"
    "github.com/chiwooi/go-swarm/repl"
    "github.com/chiwooi/go-swarm/option"
)

func main() {
    ctx := goswarm.NewContext(context.Background())

    repl.RunDemoLoop(ctx, weatherAgent, option.WithStream(true), option.WithDebug(false))
}
