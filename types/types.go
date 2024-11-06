package types

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
type AgentFunction any

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
