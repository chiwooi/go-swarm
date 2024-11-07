package main

import (
    "context"
    "encoding/json"
    "fmt"
    "testing"

    "github.com/stretchr/testify/assert"

    "github.com/openai/openai-go"

    "github.com/chiwooi/go-swarm"
    "github.com/chiwooi/go-swarm/option"
    "github.com/chiwooi/go-swarm/types"
)

var CONVERSATIONAL_EVAL_SYSTEM_PROMPT = `
You will be provided with a conversation between a user and an agent, as well as a main goal for the conversation.
Your goal is to evaluate, based on the conversation, if the agent achieves the main goal or not.

To assess whether the agent manages to achieve the main goal, consider the instructions present in the main goal, as well as the way the user responds:
is the answer satisfactory for the user or not, could the agent have done better considering the main goal?
It is possible that the user is not satisfied with the answer, but the agent still achieves the main goal because it is following the instructions provided as part of the main goal.
`

func ConversationWasSuccessful(messages interface{}) bool {
    jsonData, _ := json.Marshal(messages)
    conversation := fmt.Sprintf("CONVERSATION: %v", string(jsonData))
    result, _ := evaluateWithLLMBool(CONVERSATIONAL_EVAL_SYSTEM_PROMPT, conversation)
    return result.Value
}

func RunAndGetToolCalls(agent *types.Agent, query string) []openai.ChatCompletionMessageToolCall {
    oai := openai.NewClient()
    client := goswarm.NewSwarm(oai)

    messages := goswarm.NewMessages(openai.UserMessage(query))

    ctx := goswarm.NewContext(context.Background())
    response := client.Run(ctx, agent, messages, option.WithExecuteTools(false))
    return response.Messages[len(response.Messages)-1].(openai.ChatCompletionMessage).ToolCalls
}

func TestTriageAgentCallsCorrectFunction(t *testing.T) {
    dataSet := []struct{
        query string
        functionName string
    }{
        {"I want to make a refund!", "transferToRefunds"},
        {"I want to talk to sales.", "transferToSales"},
    }

    for _, data := range dataSet {
        toolCalls := RunAndGetToolCalls(triageAgent, data.query)

        assert.Len(t, toolCalls, 1, "Expected one tool call")
        assert.Equal(t, data.functionName, toolCalls[0].Function.Name, "Expected function name to be correct")
    }
}

func TestConversationWasSuccessful(t *testing.T) {
    dataSet := [][]map[string]string{
        {
            {"role": "user", "content": "Who is the lead singer of U2"},
            {"role": "assistant", "content": "Bono is the lead singer of U2."},
        },
        {
            {"role": "user", "content": "Hello!"},
            {"role": "assistant", "content": "Hi there! How can I assist you today?"},
            {"role": "user", "content": "I want to make a refund."},
            {"role": "tool", "tool_name": "transfer_to_refunds"},
            {"role": "user", "content": "Thank you!"},
            {"role": "assistant", "content": "You're welcome! Have a great day!"},
        },
    }

    for _, data := range dataSet {
        result := ConversationWasSuccessful(data)
        assert.Equal(t, true, result, "Expected conversation to be successful")
    }
}



