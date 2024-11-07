package main

import (
	"context"
	"fmt"
	"github.com/openai/openai-go"

	"github.com/chiwooi/go-swarm"
    "github.com/chiwooi/go-swarm/option"
    "github.com/chiwooi/go-swarm/types"
)

func GetInstructions(ctx goswarm.Context) string {
	name := ctx.GetVariable("name", "User")
	return fmt.Sprintf("You are a helpful agent. Greet the user by name (%s).", name)
}

func PrintAccountDetails(ctx goswarm.Context) string {
	if ctx.IsAnalyze() {
		ctx.SetDescription("Print the user's account details.")
		return ""
	}

	userID := ctx.GetVariable("user_id", nil)
	name := ctx.GetVariable("name", nil)
	fmt.Printf("Account Details: %d %s\n", userID, name)
	return "Success"
}

func main() {
	oai := openai.NewClient()
	client := goswarm.NewSwarm(oai)

	agent := goswarm.NewAgent(
		option.WithAgentModel("gpt-4o"),
		option.WithAgentInstructions(GetInstructions),
		option.WithAgentFunctions(PrintAccountDetails),
	)

	ctx := goswarm.NewContext(context.Background())
	ctx.SetVariables(types.ContextVariables{"name": "James", "user_id": 123})

	messages := goswarm.NewMessages(openai.UserMessage("Hi!"))
	resp := client.Run(ctx, agent, messages)

	if len(resp.Messages) > 0 {
		fmt.Println(resp.Messages[0].(openai.ChatCompletionMessage).Content)
	}

	messages = goswarm.NewMessages(openai.UserMessage("Print my account details!"))
	resp = client.Run(ctx, agent, messages)

	if len(resp.Messages) > 0 {
		fmt.Println(resp.Messages[0].(openai.ChatCompletionMessage).Content)
	}
}

