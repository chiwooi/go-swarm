package main

import (
    "context"
    "encoding/json"
    "fmt"

    "github.com/openai/openai-go"
)

// BoolEvalResult represents the structure of the response we expect from the API
type BoolEvalResult struct {
    Value  bool   `json:"value"`
    Reason string `json:"reason,omitempty"`
}

func evaluateWithLLMBool(instruction, data string) (*BoolEvalResult, error) {
    // Set up the OpenAI client
    client := openai.NewClient()

    var messages []openai.ChatCompletionMessageParamUnion

    // Added constraint statements to allow conversion of the result format to BoolEvalResult structure.
    instruction += `

Evaluate the following data and return a JSON object in the format:
{
    "value": true or false,
    "reason": "The reason for the evaluation"
}`

    messages = append(messages, openai.SystemMessage(instruction))
    messages = append(messages, openai.UserMessage(data))

    // Create the chat completion request
    req := openai.ChatCompletionNewParams{
        Model:   openai.F("gpt-4o"),
        Messages: openai.F(messages),
    }

    // Make the API call
    resp, err := client.Chat.Completions.New(context.Background(), req)
    if err != nil {
        return nil, fmt.Errorf("failed to get chat completion: %v", err)
    }

    // Assuming the response contains a message with the boolean value
    var result BoolEvalResult
    err = json.Unmarshal([]byte(resp.Choices[0].Message.Content), &result)

    return &result, nil
}
