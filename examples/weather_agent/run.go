package main

import(
    "github.com/chiwooi/go-swarm/repl"
    "github.com/chiwooi/go-swarm/option"
    "github.com/chiwooi/go-swarm/types"
)

func main() {
    repl.RunDemoLoop(weatherAgent, types.Args{}, option.WithStreamOption(true), option.WithDebugOption(false))
}
