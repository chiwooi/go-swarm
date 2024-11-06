types.go package goswarm

import (
	"github.com/openai/openai-go"

	"github.com/chiwooi/go-swarm/option"
	"github.com/chiwooi/go-swarm/types"
)

// NewAgent creates a new Agent instance with the specified name, model, and instructions.
func NewAgent(name string, opts ...option.AgentOption) *types.Agent {
	options := option.DefAgentOptions
	for _, o := range opts {
		o.ApplyOption(&options)
	}

	return &types.Agent{
		Name:              name,
		Model:             options.Model,
		Instructions:      options.Instructions,
		Functions:         options.Functions,
		ToolChoice:        options.ToolChoice,
		ParallelToolCalls: options.ParallelToolCalls,
	}
}

// NewResponse creates a new Response instance.
func NewResponse(messages []openai.ChatCompletionMessageParamUnion, agent *types.Agent, contextVariables map[string]interface{}) *types.Response {
	return &types.Response{
		Messages:         messages,
		Agent:            agent,
		ContextVariables: contextVariables,
	}
}

// NewResult creates a new Result instance.
func NewResult(value string, agent *types.Agent, contextVariables map[string]interface{}) *types.Result {
	return &types.Result{
		Value:           value,
		Agent:           agent,
		ContextVariables: contextVariables,
	}
}

func NewMessages(msg openai.ChatCompletionMessageParamUnion) []openai.ChatCompletionMessageParamUnion {
	if msg == nil {
		return []openai.ChatCompletionMessageParamUnion{}
	}
	return []openai.ChatCompletionMessageParamUnion{msg}
}