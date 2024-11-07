package goswarm_test

import (
	"context"
	"fmt"
	"github.com/openai/openai-go"
	"testing"

	"github.com/chiwooi/go-swarm"
	"github.com/chiwooi/go-swarm/types"
	"github.com/chiwooi/go-swarm/option"
)

func GetInstructions(ctx goswarm.Context) string {
	name := ctx.GetVariable("name", "User")
	return fmt.Sprintf("You are a helpful agent. Greet the user by name (%s).", name)
}

func PrintAccountDetails(ctx goswarm.Context) string {
	userID := ctx.GetVariable("user_id", 12)
	name := ctx.GetVariable("name", "")
	fmt.Printf("Account Details: %v %v\n", userID, name)

	return "Success"
}

func TestSwarm_GetChatCompletion(t *testing.T) {
	oai := openai.NewClient()
	client := goswarm.NewSwarm(oai)

	agent := goswarm.NewAgent(
		option.WithAgentModel("gpt-4o"),
		option.WithAgentInstructions(GetInstructions),
		option.WithAgentFunctions(PrintAccountDetails),
	)

    ctx := goswarm.NewContext(context.Background())
	ctx.SetVariables(types.ContextVariables{"name": "James", "user_id": 123})

	resp := client.Run(ctx, agent, goswarm.NewMessages(openai.UserMessage("Hi!")))

	if len(resp.Messages) > 0 {
		fmt.Println(resp.Messages[0].(openai.ChatCompletionMessage).Content)
	}
}
