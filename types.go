package goswarm

import (
	"github.com/openai/openai-go"
)

type Args map[string]any

func (a Args) Get(key string, defaultValue any) any {
	if value, ok := a[key]; ok {
		return value
	}
	return defaultValue
}

// AgentFunction is a type alias for functions that return either a string, an Agent, or a map.
type AgentFunction func(args Args) string

// Agent represents an agent with various attributes, including name, model, instructions, and functions.
type Agent struct {
	Name               string
	Model              string
	Instructions       interface{} // Can be either string or a function returning string
	Functions          []AgentFunction
	// openai.ChatCompletionToolChoiceOptionBehaviorNone
	// openai.ChatCompletionToolChoiceOptionBehaviorAuto
	// openai.ChatCompletionToolChoiceOptionBehaviorRequired
	ToolChoice         openai.ChatCompletionToolChoiceOptionUnionParam
	ParallelToolCalls  bool
}

// Response represents the response structure with messages, the agent that generated it, and context variables.
type Response struct {
	Messages        []openai.ChatCompletionMessageParamUnion
	Agent           *Agent
	ContextVariables Args
}

// Result encapsulates the return values for an agent function.
type Result struct {
	Value           string
	Agent           *Agent
	ContextVariables Args
}

// NewAgent creates a new Agent instance with the specified name, model, and instructions.
func NewAgent(name, model string, instructions interface{}, functions []AgentFunction) *Agent {
	return &Agent{
		Name:              name,
		Model:             model,
		Instructions:      instructions,
		Functions:         functions,
		ToolChoice:        openai.ChatCompletionToolChoiceOptionBehaviorAuto,
		ParallelToolCalls: true,
	}
}

// NewResponse creates a new Response instance.
func NewResponse(messages []openai.ChatCompletionMessageParamUnion, agent *Agent, contextVariables map[string]interface{}) *Response {
	return &Response{
		Messages:         messages,
		Agent:            agent,
		ContextVariables: contextVariables,
	}
}

// NewResult creates a new Result instance.
func NewResult(value string, agent *Agent, contextVariables map[string]interface{}) *Result {
	return &Result{
		Value:           value,
		Agent:           agent,
		ContextVariables: contextVariables,
	}
}
