package goswarm

import (
	"encoding/json"
//	"errors"
	"fmt"
	"reflect"
	"runtime"

	"github.com/chiwooi/go-swarm/option"
	"github.com/chiwooi/go-swarm/types"
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
func (s *Swarm) GetChatCompletion(ctx Context, agent *types.Agent, history []openai.ChatCompletionMessageParamUnion, modelOverride string, stream bool, debug bool) (any, error) {
	var instructions string

	ctx = NewContext(ctx)
	ctx.SetAnalyze(true)

	switch v := agent.Instructions.(type) {
	case string:
		instructions = v
	case func(Context) string:
		// if reflect.TypeOf(agent.Instructions).Kind() == reflect.Func
		instructions = v(ctx)
	default:
		return nil, fmt.Errorf("invalid instructions type: %T", v)
	}

	var messages []openai.ChatCompletionMessageParamUnion

	messages = append(messages, openai.SystemMessage(instructions))
	messages = append(messages, history...)

	if debug {
		fmt.Printf("Getting chat completion for: \n%+v\n", messages)
	}

	tools := make([]openai.ChatCompletionToolParam, len(agent.Functions))
	for i, f := range agent.Functions {
		tools[i], _ = functionToJSON(ctx, f) // Assuming FunctionToJSON is defined to convert functions to JSON
	}

	// Prepare the chat completion request
	model := agent.Model
	if modelOverride != "" {
		model = modelOverride
	}

	createParams := openai.ChatCompletionNewParams{
		Model:             openai.F(model),
		Messages:          openai.F(messages),
	}

	// set tools option if there are any functions
	if len(tools) > 0 {
		createParams.Tools = openai.F(tools)
		createParams.ToolChoice = openai.F(agent.ToolChoice)
		createParams.ParallelToolCalls = openai.F(agent.ParallelToolCalls)
	}

	if stream {
		streamOpt := openai.ChatCompletionStreamOptionsParam{
			IncludeUsage: openai.F(stream),
		}
		createParams.StreamOptions = openai.F(streamOpt)

		if debug {
			fmt.Printf("Getting chat completion tools for:\n%+v\n", tools)
		}

		return s.client.Chat.Completions.NewStreaming(ctx.GetContext(), createParams), nil
	}

	return s.client.Chat.Completions.New(ctx.GetContext(), createParams)
}

// HandleFunctionResult processes the result of a function call.
func (s *Swarm) HandleFunctionResult(result interface{}, debug bool) types.Result {
	switch res := result.(type) {
	case types.Result:
		return res
	case *types.Agent:
		return types.Result{
			Value: fmt.Sprintf(`{"assistant": "%s"}`, res.Name),
			Agent: res,
		}
	case types.Agent:
		return types.Result{
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
			return types.Result{Value: errorMessage} // Returning the error as a string
		}
		return types.Result{Value: string(value)}
	}
}

// HandleToolCalls processes tool calls from the chat completion.
func (s *Swarm) HandleToolCalls(ctx Context, toolCalls []openai.ChatCompletionMessageToolCall, functions []types.AgentFunction, debug bool) types.Response {
	ctx = NewContext(ctx)
	ctx.SetAnalyze(false)

	functionMap := make(map[string]types.AgentFunction)
	for _, f := range functions {
		fnVal := reflect.ValueOf(f)
		fnName := runtime.FuncForPC(fnVal.Pointer()).Name()
		fnName = funcNameNormalization(fnName)
		functionMap[fnName] = f
	}

	partialResponse := types.Response{
		Messages:         []openai.ChatCompletionMessageParamUnion{},
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

		// tool call 요청에 대한 함수 파라메터 값수집
		var args types.ContextVariables
		if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
			if debug {
				fmt.Printf("Failed to unmarshal arguments for tool call %s: %v\n", name, err)
			}
			continue
		}

		fmt.Printf("Calling function %s with args: %+v\n", name, args)

		rawResult := callFuncByArgs(ctx, functionMap[name], args)

		result := s.HandleFunctionResult(rawResult, debug)
		partialResponse.Messages = append(partialResponse.Messages, openai.ToolMessage(toolCall.ID, result.Value))

		if result.Agent != nil {
			partialResponse.Agent = result.Agent
		}
	}

	return partialResponse
}

// RunAndStream executes the agent and streams responses.
func (s *Swarm) RunAndStream(ctx Context, agent *types.Agent, messages []openai.ChatCompletionMessageParamUnion, opts ...option.RunOption) <-chan any {
	args := option.DefRunOptions
	for _, opt := range opts {
		opt.ApplyOption(&args)
	}

	responseChan := make(chan any)
	go func() {
		defer close(responseChan)

		ctx = NewContext(ctx)
		ctx.SetAnalyze(true)

		activeAgent := agent
		history := messages
		initLen := len(messages)

		for len(history)-initLen < args.MaxTurns {
			completionRaw, err := s.GetChatCompletion(ctx, activeAgent, history, args.Model, true, args.Debug)
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

			responseChan <- "start"
			// Handle streaming chunks here
			for stream.Next() {
				chunk := stream.Current()
				responseChan <- chunk
				acc.AddChunk(chunk)
			}
			responseChan <- "end"

			if err := stream.Err(); err != nil {
				if args.Debug {
					fmt.Println("Error in stream:", err)
				}
				return
			}
			message := acc.Choices[0].Message
			debugPrint(args.Debug, "Received completion: %+v", message)
			history = append(history, message)

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
				partialResponse := s.HandleToolCalls(ctx, toolCalls, activeAgent.Functions, args.Debug)
				history = append(history, partialResponse.Messages...)
				if partialResponse.Agent != nil {
					activeAgent = partialResponse.Agent
				}
			}
		}

		responseChan <- &types.Response{
			Messages:         history[initLen:],
			Agent:            activeAgent,
		}
	}()

	return responseChan
}

// Run executes the agent and returns the response.
func (s *Swarm) Run(ctx Context, agent *types.Agent, messages []openai.ChatCompletionMessageParamUnion, opts ...option.RunOption) *types.Response {
	args := option.DefRunOptions
	for _, opt := range opts {
		opt.ApplyOption(&args)
	}

	// modelOverride string, stream bool, debug bool, maxTurns int, executeTools bool

	if args.Stream {
		responseChan := s.RunAndStream(ctx, agent, messages, opts...)
		for response := range responseChan {
			switch v := response.(type) {
			case *types.Response:
				return v
			}
		}
		return NewResponse(messages, agent)
	}

	activeAgent := agent
	history := messages
	initLen := len(messages)

	for len(history)-initLen < args.MaxTurns {
		completionRaw, err := s.GetChatCompletion(ctx, activeAgent, history, args.Model, false, args.Debug)
		if err != nil {
			if args.Debug {
				fmt.Println("Error getting chat completion:", err)
			}
			break
		}
		completion := completionRaw.(*openai.ChatCompletion)

		message := completion.Choices[0].Message
		if args.Debug {
			fmt.Printf("Received completion: %+v\n", message)
		}
		// message.Sender = activeAgent.Name
		history = append(history, message)

		if len(message.ToolCalls) == 0 || !args.ExecuteTools {
			if args.Debug {
				fmt.Println("Ending turn.")
			}
			break
		}

		partialResponse := s.HandleToolCalls(ctx, message.ToolCalls, activeAgent.Functions, args.Debug)
		history = append(history, partialResponse.Messages...)
		if partialResponse.Agent != nil {
			activeAgent = partialResponse.Agent
		}
	}

	return &types.Response{
		Messages:         history[initLen:],
		Agent:            activeAgent,
	}
}
