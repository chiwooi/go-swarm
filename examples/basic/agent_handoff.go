package main

import (
    "context"
    "fmt"
    "github.com/openai/openai-go"

    "github.com/chiwooi/go-swarm"
    "github.com/chiwooi/go-swarm/option"
    "github.com/chiwooi/go-swarm/types"
)

func main() {
    oai := openai.NewClient()
    client := goswarm.NewSwarm(oai)

    englishAgent := goswarm.NewAgent(
        option.WithAgentName("English Agent"), 
        option.WithAgentInstructions("You only speak English."),
    )
    spanishAgent := goswarm.NewAgent(
        option.WithAgentName("Spanish Agent"), 
        option.WithAgentInstructions("You only speak Spanish."),
    )

    transferToSpanishAgent := func(ctx goswarm.Context) *types.Agent {
        if ctx.IsAnalyze() {
            ctx.SetDescription("Transfer spanish speaking users immediately.")
            return spanishAgent
        }

        return spanishAgent
    }

    englishAgent.Functions = append(englishAgent.Functions, transferToSpanishAgent)

    ctx := goswarm.NewContext(context.Background())

    messages := goswarm.NewMessages(openai.UserMessage("Hola. ¿Como estás?"))
    resp := client.Run(ctx, englishAgent, messages)

    if len(resp.Messages) > 0 {
        fmt.Println(resp.Messages[0].(openai.ChatCompletionMessage).Content)
    }
}
