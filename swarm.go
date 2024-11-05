package goswarm

import (
	"context"
	"encoding/json"
//	"errors"
	"fmt"
	"reflect"
	"runtime"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/packages/ssestream"
)

var __CTX_VARS_NAME__ = "context_variables"

// Swarm represents a collection of agents that interact with OpenAI's API.
type Swarm struct {
	client *openai.Client
}

// NewSwarm initializes a Swarm with an optional OpenAI client.
func NewSwarm(client *openai.Client) *Swarm {
	if client == nil {
		client = openai.NewClient() // Initialize a new client if none is provided
	}
	return &Swarm{client: client}
}

// GetChatCompletion retrieves chat completions from the OpenAI API.
// - stream
//   true  : *ssestream.Stream[ChatCompletionChunk]
//   false : *openai.ChatCompletion
func (s *Swarm) GetChatCompletion(agent Agent, history []openai.ChatCompletionMessageParamUnion, contextVariables Args, modelOverride string, stream bool, debug bool) (any, error) {
	var instructions string

	ctx := context.Background()

	switch v := agent.Instructions.(type) {
	case string:
		instructions = v
	case func(Args) string:
		// if reflect.TypeOf(agent.Instructions).Kind() == reflect.Func
		instructions = v(contextVariables)
	default:
		return nil, fmt.Errorf("invalid instructions type: %T", v)
	}

	var messages []openai.ChatCompletionMessageParamUnion

	messages = append(messages, openai.SystemMessage(instructions))
	messages = append(messages, history...)

	if debug {
		fmt.Println("Getting chat completion for:", messages)
	}

	tools := make([]openai.ChatCompletionToolParam, len(agent.Functions))
	for i, f := range agent.Functions {
		tools[i], _ = functionToJSON(f) // Assuming FunctionToJSON is defined to convert functions to JSON
	}

	// Remove context variables from tools
	for i := range tools {
		params := tools[i].Function.Value.Parameters.Value
		if params == nil {
			continue
		}
		if propertys, ok := params["properties"].(map[string]interface{}); ok {
			delete(propertys, __CTX_VARS_NAME__)
		}

		// 필수 항목중 context_variables 제거		
		required := params["required"].([]string)
		for j, req := range required {
			if req == __CTX_VARS_NAME__ {
				params["required"] = append(required[:j], required[j+1:]...) // Remove from required
				break
			}
		}
	}

	// Prepare the chat completion request
	model := agent.Model
	if modelOverride != "" {
		model = modelOverride
	}

	createParams := openai.ChatCompletionNewParams{
		Model:             openai.F(model),
		Messages:          openai.F(messages),
		Tools:             openai.F(tools),
		ToolChoice:        openai.F(agent.ToolChoice),
		// StreamOptions:     openai.F(streamOpt),
		// ParallelToolCalls: openai.F(agent.ParallelToolCalls),
	}

	if len(tools) > 0 {
		createParams.ParallelToolCalls = openai.F(agent.ParallelToolCalls)
	}

	if stream {
		streamOpt := openai.ChatCompletionStreamOptionsParam{
			IncludeUsage: openai.F(stream),
		}
		createParams.StreamOptions = openai.F(streamOpt)

		return s.client.Chat.Completions.NewStreaming(ctx, createParams), nil
	}

	return s.client.Chat.Completions.New(ctx, createParams)
}

// HandleFunctionResult processes the result of a function call.
func (s *Swarm) HandleFunctionResult(result interface{}, debug bool) Result {
	switch res := result.(type) {
	case Result:
		return res
	case Agent:
		return Result{
			Value: fmt.Sprintf(`{"assistant": "%s"}`, res.Name),
			Agent: &res,
		}
	default:
		value, err := json.Marshal(result)
		if err != nil {
			errorMessage := fmt.Sprintf("Failed to cast response to string: %v. Ensure agent functions return a string or Result object. Error: %v", result, err)
			if debug {
				fmt.Println(errorMessage)
			}
			return Result{Value: errorMessage} // Returning the error as a string
		}
		return Result{Value: string(value)}
	}
}

// HandleToolCalls processes tool calls from the chat completion.
func (s *Swarm) HandleToolCalls(toolCalls []openai.ChatCompletionMessageToolCall, functions []AgentFunction, contextVariables Args, debug bool) Response {
	functionMap := make(map[string]AgentFunction)
	for _, f := range functions {
		fnVal := reflect.ValueOf(f)
		name := runtime.FuncForPC(fnVal.Pointer()).Name()
		functionMap[name] = f
	}

	partialResponse := Response{
		Messages:         []openai.ChatCompletionMessageParamUnion{},
		ContextVariables: make(Args),
	}

	for _, toolCall := range toolCalls {
		name := toolCall.Function.Name
		if _, found := functionMap[name]; !found {
			if debug {
				fmt.Printf("Tool %s not found in function map.\n", name)
			}
			// ToolName:   name,
			partialResponse.Messages = append(partialResponse.Messages, 
				openai.ToolMessage(toolCall.ID, fmt.Sprintf("Error: Tool %s not found.", name)),
			)
			continue
		}

		var args Args
		if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
			if debug {
				fmt.Printf("Failed to unmarshal arguments for tool call %s: %v\n", name, err)
			}
			continue
		}

		// 호출함수 정의에 __CTX_VARS_NAME__ 인자가 있는지 확인
		if hasArgInFunc(functionMap[name], __CTX_VARS_NAME__) {
			args[__CTX_VARS_NAME__] = contextVariables
		}

		rawResult := callFuncByArgs(functionMap[name], args)

		result := s.HandleFunctionResult(rawResult, debug)
		partialResponse.Messages = append(partialResponse.Messages, openai.ToolMessage(toolCall.ID, result.Value))

		for k, v := range result.ContextVariables {
			partialResponse.ContextVariables[k] = v
		}
		if result.Agent != nil {
			partialResponse.Agent = result.Agent
		}
	}

	return partialResponse
}

// RunAndStream executes the agent and streams responses.
func (s *Swarm) RunAndStream(agent Agent, messages []openai.ChatCompletionMessageParamUnion, contextVariables Args, opts ...RunOption) <-chan Response {
	args := DefRunOptions
	for _, opt := range opts {
		opt.ApplyOption(&args)
	}

	responseChan := make(chan Response)
	go func() {
		defer close(responseChan)

		activeAgent := agent
		history := messages
		initLen := len(messages)

		for len(history)-initLen < args.MaxTurns {
			completionRaw, err := s.GetChatCompletion(activeAgent, history, contextVariables, args.Model, true, args.Debug)
			if err != nil {
				if args.Debug {
					fmt.Println("Error getting chat completion:", err)
				}
				return
			}

			stream, ok := completionRaw.(*ssestream.Stream[openai.ChatCompletionChunk])
			if !ok {
				if args.Debug {
					fmt.Println("Error casting stream.")
				}
				return
			}

			acc := openai.ChatCompletionAccumulator{}

			// Handle streaming chunks here
			for stream.Next() {
				chunk := stream.Current()
				acc.AddChunk(chunk)
			}

			if err := stream.Err(); err != nil {
				if args.Debug {
					fmt.Println("Error in stream:", err)
				}
				return
			}

			message := acc.Choices[0].Message
			if len(message.ToolCalls) == 0 {
				if args.Debug {
					fmt.Println("Ending turn.")
				}
				break
			}

			if args.ExecuteTools {
				toolCalls := []openai.ChatCompletionMessageToolCall{}
				for _, toolCall := range message.ToolCalls {
					toolCalls = append(toolCalls, toolCall)
				}
				partialResponse := s.HandleToolCalls(toolCalls, activeAgent.Functions, contextVariables, args.Debug)
				history = append(history, partialResponse.Messages...)
				for k, v := range partialResponse.ContextVariables {
					contextVariables[k] = v
				}
				if partialResponse.Agent != nil {
					activeAgent = *partialResponse.Agent
				}
			}
		}

		responseChan <- Response{
			Messages:         history[initLen:],
			Agent:            &activeAgent,
			ContextVariables: contextVariables,
		}
	}()

	return responseChan
}

// Run executes the agent and returns the response.
func (s *Swarm) Run(agent Agent, messages []openai.ChatCompletionMessageParamUnion, contextVariables Args, opts ...RunOption) Response {
	args := DefRunOptions

	for _, opt := range opts {
		opt.ApplyOption(&args)
	}

	// modelOverride string, stream bool, debug bool, maxTurns int, executeTools bool

	if args.Stream {
		responseChan := s.RunAndStream(agent, messages, contextVariables, opts...)
		return <-responseChan
	}

	activeAgent := agent
	history := messages
	initLen := len(messages)

	for len(history)-initLen < args.MaxTurns {
		completionRaw, err := s.GetChatCompletion(activeAgent, history, contextVariables, args.Model, false, args.Debug)
		if err != nil {
			if args.Debug {
				fmt.Println("Error getting chat completion:", err)
			}
			break
		}
		completion := completionRaw.(*openai.ChatCompletion)

		message := completion.Choices[0].Message
		if args.Debug {
			fmt.Println("Received completion:", message)
		}
		// message.Sender = activeAgent.Name
		history = append(history, message)

		if len(message.ToolCalls) == 0 || !args.ExecuteTools {
			if args.Debug {
				fmt.Println("Ending turn.")
			}
			break
		}

		partialResponse := s.HandleToolCalls(message.ToolCalls, activeAgent.Functions, contextVariables, args.Debug)
		history = append(history, partialResponse.Messages...)
		for k, v := range partialResponse.ContextVariables {
			contextVariables[k] = v
		}
		if partialResponse.Agent != nil {
			activeAgent = *partialResponse.Agent
		}
	}

	return Response{
		Messages:         history[initLen:],
		Agent:            &activeAgent,
		ContextVariables: contextVariables,
	}
}
