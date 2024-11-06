// https://github.com/openai/swarm/blob/main/examples/basic/function_calling.py
package main

import (
	"fmt"
	"github.com/openai/openai-go"

	"github.com/chiwooi/go-swarm"
	"github.com/chiwooi/go-swarm/option"
	"github.com/chiwooi/go-swarm/types"
)

type GetWeatherArgs struct {
	Loc string `json:"loc" desc:"The location to get the weather for." required:"true"`
}

func GetWeather(ctx goswarm.Context, args GetWeatherArgs) string {
	if ctx.IsAnalyze() {
		ctx.SetDescription("Get the weather for a location.")
		return ""
	}

	return "{'temp':67, 'unit':'F'}"
}

func main() {
	oai := openai.NewClient()
	client := goswarm.NewSwarm(oai)

	agent := goswarm.NewAgent("Agent",
		option.WithAgentInstructions("You are a helpful agent."),
		option.WithAgentFunctions(GetWeather),
	)

	messages := goswarm.NewMessages(openai.UserMessage("What's the weather in NYC?"))
	resp := client.Run(agent, messages, types.Args{})

	if len(resp.Messages) > 0 {
		fmt.Println(resp.Messages[0].(openai.ChatCompletionMessage).Content)
	}
}
