package main

import (
	"fmt"
	"github.com/openai/openai-go"

	"github.com/chiwooi/go-swarm"
    "github.com/chiwooi/go-swarm/option"
    "github.com/chiwooi/go-swarm/types"
)

func GetInstructions(ctx goswarm.Context) string {
	name := ctx.GetArg("name", "User")
	return fmt.Sprintf("You are a helpful agent. Greet the user by name (%s).", name)
}

func PrintAccountDetails(ctx goswarm.Context) string {
	if ctx.IsAnalyze() {
		ctx.SetDescription("Print the user's account details.")
		return ""
	}

	userID := ctx.GetArg("user_id", nil)
	name := ctx.GetArg("name", nil)
	fmt.Printf("Account Details: %d %s\n", userID, name)
	return "Success"
}

func main() {
	oai := openai.NewClient()
	client := goswarm.NewSwarm(oai)

	agent := goswarm.NewAgent("Agent",
		option.WithAgentModel("gpt-4o"),
		option.WithAgentInstructions(GetInstructions),
		option.WithAgentFunctions(PrintAccountDetails),
	)

	contextVariables := types.Args{"name": "James", "user_id": 123}

	messages := goswarm.NewMessages(openai.UserMessage("Hi!"))
	resp := client.Run(agent, messages, contextVariables)

	if len(resp.Messages) > 0 {
		fmt.Println(resp.Messages[0].(openai.ChatCompletionMessage).Content)
	}

	messages = goswarm.NewMessages(openai.UserMessage("Print my account details!"))
	resp = client.Run(agent, messages, contextVariables)

	if len(resp.Messages) > 0 {
		fmt.Println(resp.Messages[0].(openai.ChatCompletionMessage).Content)
	}
}

