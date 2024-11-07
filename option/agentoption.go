package option

import (
	"reflect"

	"github.com/chiwooi/go-swarm/types"
	"github.com/openai/openai-go"
)

type AgentOption interface {
   ApplyOption(opts *AgentOptions)
}

type AgentOptions struct {
	Name              string
	Model             string
	Instructions      any
	Functions         []types.AgentFunction
	ToolChoice        openai.ChatCompletionToolChoiceOptionUnionParam
	ParallelToolCalls bool
}

var DefAgentOptions = AgentOptions{
	Name:              "Agent",
   Model:             "gpt-4o",
   Instructions:      "You are a helpful agent.",
   ToolChoice:        openai.ChatCompletionToolChoiceOptionBehaviorAuto,
   ParallelToolCalls: true,
}

// set the model for the agent.

type AgentNameOption string

func (o AgentNameOption) ApplyOption(opts *AgentOptions) {
   opts.Model = string(o)
}

func WithAgentName(name string) AgentNameOption {
   return AgentNameOption(name)
}

// set the model for the agent.

type AgentModelOption string

func (o AgentModelOption) ApplyOption(opts *AgentOptions) {
   opts.Model = string(o)
}

func WithAgentModel(model string) AgentModelOption {
   return AgentModelOption(model)
}

// set the instructions for the agent.

type AgentInstructionsOption struct {
	instructions any
}

func (o AgentInstructionsOption) ApplyOption(opts *AgentOptions) {
   opts.Instructions = o.instructions
}

func WithAgentInstructions(inst any) AgentInstructionsOption {
   return AgentInstructionsOption{inst}
}

// set the tool choice for the agent.

type AgentToolChoiceOption string
const (
	ToolChoiceOptionNone     AgentToolChoiceOption = "none"
	ToolChoiceOptionAuto     AgentToolChoiceOption = "auto"
	ToolChoiceOptionRequired AgentToolChoiceOption = "required"
)

func (o AgentToolChoiceOption) ApplyOption(opts *AgentOptions) {
	switch o {
	case ToolChoiceOptionNone:
	   opts.ToolChoice = openai.ChatCompletionToolChoiceOptionBehaviorNone
	case ToolChoiceOptionAuto:
	   opts.ToolChoice = openai.ChatCompletionToolChoiceOptionBehaviorAuto
	case ToolChoiceOptionRequired:
	   opts.ToolChoice = openai.ChatCompletionToolChoiceOptionBehaviorRequired
	}
}

func WithAgentToolChoice(toolChoice AgentToolChoiceOption) AgentToolChoiceOption {
   return AgentToolChoiceOption(toolChoice)
}


// set the functions for the agent.

type AgentFunctionsOption struct {
	fns []types.AgentFunction
}

func (o AgentFunctionsOption) ApplyOption(opts *AgentOptions) {
	opts.Functions = append(opts.Functions, o.fns...)
}

func WithAgentFunctions(fn ...types.AgentFunction) AgentFunctionsOption {
	fns := AgentFunctionsOption{}

	for _, f := range fn {
		if reflect.TypeOf(f).Kind() != reflect.Func {
			panic("provided value is not a function")
		}
		fns.fns = append(fns.fns, f)
	}

	return fns
}

// set the parallel tool calls for the agent.

type AgentParallelToolCallsOption bool

func (o AgentParallelToolCallsOption) ApplyOption(opts *AgentOptions) {
   opts.ParallelToolCalls = bool(o)
}

func WithAgentParallelToolCalls(flag bool) AgentParallelToolCallsOption {
   return AgentParallelToolCallsOption(flag)
}
