// https://github.com/openai/swarm/blob/main/examples/basic/function_calling.py
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/openai/openai-go"

	"github.com/chiwooi/go-swarm"
	"github.com/chiwooi/go-swarm/option"
	"github.com/chiwooi/go-swarm/types"
)

func PrettyPrintMessage(messages []openai.ChatCompletionMessageParamUnion) {
    for _, message := range messages {
        msg := message.(openai.ChatCompletionMessage)
        fmt.Printf("%v: %v", "sender", msg.Content)
    }
}

func main() {
    oai := openai.NewClient()
    client := goswarm.NewSwarm(oai)

    agent := goswarm.NewAgent("Agent",
        option.WithAgentInstructions("You are a helpful agent."),
    )

    messages := goswarm.NewMessages(nil)
    reader := bufio.NewReader(os.Stdin)

    for {
        fmt.Printf("\n> ")
        userInput, _ := reader.ReadString('\n')
        userInput = strings.ReplaceAll(userInput, "\n", "")
        messages = append(messages, openai.UserMessage(userInput))

        response := client.Run(agent, messages, types.Args{}, option.WithDebugOption(false))
        messages = response.Messages
        agent = response.Agent
        PrettyPrintMessage(messages)
    }
}



