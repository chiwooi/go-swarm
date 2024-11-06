package goswarm_test

import (
	"fmt"
	"github.com/openai/openai-go"
	"testing"

	"github.com/chiwooi/go-swarm"
	"github.com/chiwooi/go-swarm/types"
	"github.com/chiwooi/go-swarm/option"
)

func GetInstructions(args types.Args) string {
	name := args.Get("name", "User")
	return fmt.Sprintf("You are a helpful agent. Greet the user by name (%s).", name)
}

func PrintAccountDetails(args types.Args) string {
	userID := args.Get("user_id", nil)
	name := args.Get("name", nil)
	fmt.Printf("Account Details: %s %s\n", userID, name)
	return "Success"
}

func TestSwarm_GetChatCompletion(t *testing.T) {
	oai := openai.NewClient()

	agent := goswarm.NewAgent("Agent",
		option.WithAgentModel("gpt-4o"),
		option.WithAgentInstructions(GetInstructions),
		option.WithAgentFunctions(PrintAccountDetails),
	)

	client := goswarm.NewSwarm(oai)
	resp := client.Run(agent, goswarm.NewMessages(openai.UserMessage("Hi!")), types.Args{"name": "James", "user_id": 123})

	if len(resp.Messages) > 0 {
		fmt.Println(resp.Messages[0].(openai.ChatCompletionMessage).Content)
	}
}

