package main

import (
    "context"
    "fmt"
    "github.com/openai/openai-go"

    "github.com/chiwooi/go-swarm"
    "github.com/chiwooi/go-swarm/option"
)

func main() {
    oai := openai.NewClient()
    client := goswarm.NewSwarm(oai)

    agent := goswarm.NewAgent(
        option.WithAgentInstructions("You are a helpful agent."),
    )

    ctx := goswarm.NewContext(context.Background())

    messages := goswarm.NewMessages(openai.UserMessage("Hi!"))
    resp := client.Run(ctx, agent, messages)

    if len(resp.Messages) > 0 {
        fmt.Println(resp.Messages[0].(openai.ChatCompletionMessage).Content)
    }
}
