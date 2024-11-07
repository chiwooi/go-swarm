package main

import (
    "context"
    "testing"
    "github.com/stretchr/testify/assert"

    "github.com/openai/openai-go"

    "github.com/chiwooi/go-swarm"
    "github.com/chiwooi/go-swarm/option"
    "github.com/chiwooi/go-swarm/types"
)

func RunAndGetToolCalls(agent *types.Agent, query string) []openai.ChatCompletionMessageToolCall {
    ctx := goswarm.NewContext(context.Background())

    messages := goswarm.NewMessages(openai.UserMessage(query))
    response := client.Run(ctx, agent, messages, option.WithExecuteTools(false))

    return response.Messages[len(response.Messages)-1].(openai.ChatCompletionMessage).ToolCalls
}

func TestCallsWeatherWhenAsked(t *testing.T) {
    querys := []string{
        "What's the weather in NYC?",
        "Tell me the weather in London.",
        "Do I need an umbrella today? I'm in chicago.",
    }

    for _, query := range querys {
        toolCalls := RunAndGetToolCalls(weatherAgent, query)

        assert.Equal(t, 1, len(toolCalls), "Expected 1 tool call, got %d", len(toolCalls))
        assert.Equal(t, "GetWeather", toolCalls[0].Function.Name, "Expected tool call to GetWeather, got %s", toolCalls[0].Function.Name)
    }
}

func TestDoesNotCallWeatherWhenNotAsked(t *testing.T) {
    querys := []string{
        "Who's the president of the United States?",
        "What is the time right now?",
        "Hi!",
    }

    for _, query := range querys {
        toolCalls := RunAndGetToolCalls(weatherAgent, query)

        assert.Equal(t, 0, len(toolCalls), "Expected 0 tool calls, got %d", len(toolCalls))
    }
}
