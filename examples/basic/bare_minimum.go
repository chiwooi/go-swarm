package main

import (
    "fmt"
    "github.com/openai/openai-go"

    "github.com/chiwooi/go-swarm"
    "github.com/chiwooi/go-swarm/option"
    "github.com/chiwooi/go-swarm/types"
)

func main() {
    oai := openai.NewClient()
    client := goswarm.NewSwarm(oai)

    agent := goswarm.NewAgent("Agent", option.WithAgentInstructions("You are a helpful agent."))

    messages := goswarm.NewMessages(openai.UserMessage("Hi!"))
    resp := client.Run(agent, messages, types.Args{})

    if len(resp.Messages) > 0 {
        fmt.Println(resp.Messages[0].(openai.ChatCompletionMessage).Content)
    }
}
