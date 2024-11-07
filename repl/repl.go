package repl

import (
    "fmt"
    "bufio"
    "os"
    "strings"

    "github.com/openai/openai-go"

    "github.com/chiwooi/go-swarm"
    "github.com/chiwooi/go-swarm/option"
    "github.com/chiwooi/go-swarm/types"
)

func ProcessAndPrintStreamingResponse(response <-chan any) *types.Response {
    var content string
    var lastSender string

    acc := openai.ChatCompletionAccumulator{}

    for chunk := range response {
        // if msg.Sender != "" {
        //     lastSender = msg.Sender
        // }

        switch v := chunk.(type) {
        case *types.Response:
            return v
        case openai.ChatCompletionChunk:
            acc.AddChunk(v)
            if len(v.Choices) == 0 {
                break
            }

            if v.Choices[0].Delta.Content != "" {
                if lastSender != "" {
                    fmt.Printf("\033[94m%s:\033[0m ", lastSender)
                    lastSender = ""
                }
                fmt.Print(v.Choices[0].Delta.Content)
                content += v.Choices[0].Delta.Content
            }

            if len(v.Choices[0].Delta.ToolCalls) > 0 {
                for _, toolCall := range v.Choices[0].Delta.ToolCalls {
                    f := toolCall.Function
                    name := f.Name
                    if name == "" {
                        continue
                    }
                    fmt.Printf("\033[94m%s: \033[95m%s\033[0m()\n", lastSender, name)
                }
            }

            if content, ok := acc.JustFinishedContent(); ok {
                fmt.Printf("\033[94m%s:\033[0m ", lastSender)
                fmt.Print(content)
            }
            if tool, ok := acc.JustFinishedToolCall(); ok {
                fmt.Println("Tool call stream finished:", tool.Index, tool.Name, tool.Arguments)
            }
            if refusal, ok := acc.JustFinishedRefusal(); ok {
                fmt.Println("Refusal stream finished:", refusal)
            }
        case string:
            switch v {
            case "start":
            case "end":
                fmt.Println()
                content = ""
            }
        }
    }

    return nil
}

func convertToMessage(param openai.ChatCompletionMessageParamUnion) openai.ChatCompletionMessage {
    switch v := param.(type) {
    case openai.ChatCompletionMessage:
        return v
    case openai.ChatCompletionMessageParam:
        return openai.ChatCompletionMessage{
            Role:         openai.ChatCompletionMessageRole(v.Role.Value),
            Content:      v.Content.String(),
            FunctionCall: v.FunctionCall.Value.(openai.ChatCompletionMessageFunctionCall),
            ToolCalls:    v.ToolCalls.Value.([]openai.ChatCompletionMessageToolCall),
        }
    }
    return openai.ChatCompletionMessage{}
}

func PrettyPrintMessages(messages []openai.ChatCompletionMessageParamUnion) {
    for _, message := range messages {
        msg := convertToMessage(message)
        if msg.Role != "assistant" {
            continue
        }

        // print agent name in blue
        fmt.Printf("\033[94m%s\033[0m: ", "sender")

        // print response, if any
        if msg.Content != "" {
            fmt.Println(msg.Content)
        }

        // print tool calls in purple, if any
        toolCalls := msg.ToolCalls
        if len(toolCalls) > 1 {
            fmt.Println()
        }
        for _, toolCall := range toolCalls {
            f := toolCall.Function
            name, args := f.Name, f.Arguments
            argStr := strings.ReplaceAll(fmt.Sprintf("%v", args), ":", "=")
            fmt.Printf("\033[95m%s\033[0m(%s)\n", name, argStr[1:len(argStr)-1])
        }
    }
}


func RunDemoLoop(ctx goswarm.Context, startAgent *types.Agent, opts ...option.RunOption) {
    args := option.DefRunOptions
    for _, opt := range opts {
        opt.ApplyOption(&args)
    }

    oai := openai.NewClient()
    client := goswarm.NewSwarm(oai)

    fmt.Println("Starting Swarm CLI 🐝")

    reader := bufio.NewReader(os.Stdin)
    messages := goswarm.NewMessages(nil)
    agent := startAgent

    for {
        fmt.Printf("\033[90mUser\033[0m: ")
        userInput, _ := reader.ReadString('\n')
        userInput = strings.ReplaceAll(userInput, "\n", "")

        messages = append(messages, openai.UserMessage(userInput))

        var response *types.Response


        if args.Stream {
            responseChan := client.RunAndStream(ctx, agent, messages, opts...)
            response = ProcessAndPrintStreamingResponse(responseChan)
        } else {
            response = client.Run(ctx, agent, messages, opts...)
            PrettyPrintMessages(response.Messages)
        }

        if response != nil {
            messages = append(messages, response.Messages...)
            agent = response.Agent
        }
    }
}
