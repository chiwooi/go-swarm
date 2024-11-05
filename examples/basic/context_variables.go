package main

import (
	"fmt"
	"github.com/openai/openai-go"

	"github.com/chiwooi/go-swarm"
)

func GetInstructions(args goswarm.Args) string {
	name := args.Get("name", "User")
	return fmt.Sprintf("You are a helpful agent. Greet the user by name (%s).", name)
}

func PrintAccountDetails(args goswarm.Args) string {
	userID := args.Get("user_id", nil)
	name := args.Get("name", nil)
	fmt.Printf("Account Details: %s %s\n", userID, name)
	return "Success"
}

func main2() {
	oai := openai.NewClient()

	agent := goswarm.Agent{
		Model: "gpt-4o",
	    Name: "Agent",
	    Instructions: GetInstructions,
	    Functions: []goswarm.AgentFunction{PrintAccountDetails},
	}

	client := goswarm.NewSwarm(oai)

	messages := []openai.ChatCompletionMessageParamUnion{}
	messages = append(messages, openai.UserMessage("Hi!"))

	resp := client.Run(agent, messages, goswarm.Args{"name": "James", "user_id": 123})

	if len(resp.Messages) > 0 {
		fmt.Println(resp.Messages[0].(openai.ChatCompletionMessage).Content)
	}
}

