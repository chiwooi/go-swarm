package main

import (
	"fmt"
	"github.com/openai/openai-go"

	"github.com/chiwooi/go-swarm"
)


func GetWeather(loc string, args goswarm.Args) string {
	return "{'temp':67, 'unit':'F'}"
}

func main() {
	oai := openai.NewClient()

	agent := goswarm.Agent{
		Model: "gpt-4o",
	    Name: "Agent",
	    Instructions: "You are a helpful agent.",
	    Functions: []goswarm.AgentFunction{GetWeather},
	}

	client := goswarm.NewSwarm(oai)

	messages := []openai.ChatCompletionMessageParamUnion{}
	messages = append(messages, openai.UserMessage("What's the weather in NYC?"))
	resp := client.Run(agent, messages, goswarm.Args{"name": "James", "user_id": 123})

	if len(resp.Messages) > 0 {
		fmt.Println(resp.Messages[0].(openai.ChatCompletionMessage).Content)
	}
}

