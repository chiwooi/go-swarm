package main

import(
    "context"

    "github.com/chiwooi/go-swarm"
    "github.com/chiwooi/go-swarm/repl"
)

func main() {
    ctx := goswarm.NewContext(context.Background())

    repl.RunDemoLoop(ctx, triageAgent)
}
