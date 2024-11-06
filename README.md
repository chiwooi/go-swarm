# go-swarm

![Swarm Logo](assets/logo.png)

# Swarm (experimental, educational)

An educational framework exploring ergonomic, lightweight multi-agent orchestration.

> [!WARNING]
> Swarm is currently an experimental sample framework intended to explore ergonomic interfaces for multi-agent systems. It is not intended to be used in production, and therefore has no official support. (This also means we will not be reviewing PRs or issues!)
>
> The primary goal of Swarm is to showcase the handoff & routines patterns explored in the [Orchestrating Agents: Handoffs & Routines](https://cookbook.openai.com/examples/orchestrating_agents) cookbook. It is not meant as a standalone library, and is primarily for educational purposes.

## Install

Requires golang 1.21+

```shell
import "github.com/chiwooi/go-swarm"
```

## Usage

```go
package main

import (
    "fmt"
    "github.com/openai/openai-go"

    "github.com/chiwooi/go-swarm"
    "github.com/chiwooi/go-swarm/option"
    "github.com/chiwooi/go-swarm/types"
)

func main() {
    oai := openai.NewClient()
    client := goswarm.NewSwarm(oai)

    agentA := goswarm.NewAgent("Agent A", option.WithAgentInstructions("You are a helpful agent."))
    agentB := goswarm.NewAgent("Agent B", option.WithAgentInstructions("Only speak in Haikus."))

    transferToAgentB := func(ctx goswarm.Context) *types.Agent {
        if ctx.IsAnalyze() {
            ctx.SetDescription("transfer to agent b.")
            return nil
        }
        return agentB
    }

    agentA.Functions = append(agentA.Functions, transferToAgentB)

    messages := goswarm.NewMessages(openai.UserMessage("I want to talk to agent B."))
    resp := client.Run(agentA, messages, types.Args{})

    if len(resp.Messages) > 0 {
        fmt.Println(resp.Messages[0].(openai.ChatCompletionMessage).Content)
    }
}

```

```
Hope glimmers brightly,
New paths converge gracefully,
What can I assist?
```

## Table of Contents

- [Overview](#overview)
- [Examples](#examples)
- [Documentation](#documentation)
  - [Running Swarm](#running-swarm)
  - [Agents](#agents)
  - [Functions](#functions)
  - [Streaming](#streaming)
- [Evaluations](#evaluations)
- [Utils](#utils)

# Overview

Swarm focuses on making agent **coordination** and **execution** lightweight, highly controllable, and easily testable.

It accomplishes this through two primitive abstractions: `Agent`s and **handoffs**. An `Agent` encompasses `instructions` and `tools`, and can at any point choose to hand off a conversation to another `Agent`.

These primitives are powerful enough to express rich dynamics between tools and networks of agents, allowing you to build scalable, real-world solutions while avoiding a steep learning curve.

> [!NOTE]
> Swarm Agents are not related to Assistants in the Assistants API. They are named similarly for convenience, but are otherwise completely unrelated. Swarm is entirely powered by the Chat Completions API and is hence stateless between calls.

## Why Swarm

Swarm explores patterns that are lightweight, scalable, and highly customizable by design. Approaches similar to Swarm are best suited for situations dealing with a large number of independent capabilities and instructions that are difficult to encode into a single prompt.

The Assistants API is a great option for developers looking for fully-hosted threads and built in memory management and retrieval. However, Swarm is an educational resource for developers curious to learn about multi-agent orchestration. Swarm runs (almost) entirely on the client and, much like the Chat Completions API, does not store state between calls.

# Examples

Check out `/examples` for inspiration! Learn more about each one in its README.

- [`basic`](examples/basic): Simple examples of fundamentals like setup, function calling, handoffs, and context variables
- [`triage_agent`](examples/triage_agent): Simple example of setting up a basic triage step to hand off to the right agent
- [`weather_agent`](examples/weather_agent): Simple example of function calling
- [`airline`](examples/airline): A multi-agent setup for handling different customer service requests in an airline context.
- [`support_bot`](examples/support_bot): A customer service bot which includes a user interface agent and a help center agent with several tools
- [`personal_shopper`](examples/personal_shopper): A personal shopping agent that can help with making sales and refunding orders

# Documentation

![Swarm Diagram](assets/swarm_diagram.png)

## Running Swarm

Start by instantiating a Swarm client (which internally just instantiates an `OpenAI` client).

```go
import (
    "fmt"
    "github.com/openai/openai-go"

    "github.com/chiwooi/go-swarm"
    "github.com/chiwooi/go-swarm/option"
    "github.com/chiwooi/go-swarm/types"
)

oai := openai.NewClient()
client := goswarm.NewSwarm(oai)

```

### `client.Run()`

Swarm's `run()` function is analogous to the `chat.completions.create()` function in the Chat Completions API – it takes `messages` and returns `messages` and saves no state between calls. Importantly, however, it also handles Agent function execution, hand-offs, context variable references, and can take multiple turns before returning to the user.

At its core, Swarm's `client.run()` implements the following loop:

1. Get a completion from the current Agent
2. Execute tool calls and append results
3. Switch Agent if necessary
4. Update context variables, if necessary
5. If no new function calls, return

#### Arguments

| Argument              | Type    | Description                                                                                                                                            | Default        |
| --------------------- | ------- | ------------------------------------------------------------------------------------------------------------------------------------------------------ | -------------- |
| **agent**             | `Agent` | The (initial) agent to be called.                                                                                                                      | (required)     |
| **messages**          | `List`  | A list of message objects, identical to [Chat Completions `messages`](https://platform.openai.com/docs/api-reference/chat/create#chat-create-messages) | (required)     |
| **contextVariables** | `dict`  | A dictionary of additional context variables, available to functions and Agent instructions                                                            | `{}`           |
| **option.WithMaxTurnsOption()**         | `int`   | The maximum number of conversational turns allowed                                                                                                     | `float("inf")` |
| **option.WithModel()**    | `str`   | An optional string to override the model being used by an Agent                                                                                        | `None`         |
| **option.WithExecuteToolsOption()**     | `bool`  | If `False`, interrupt execution and immediately returns `tool_calls` message when an Agent tries to call a function                                    | `True`         |
| **option.WithStreamOption()**            | `bool`  | If `True`, enables streaming responses                                                                                                                 | `False`        |
| **option.WithDebugOption()**             | `bool`  | If `True`, enables debug logging                                                                                                                       | `False`        |

Once `client.run()` is finished (after potentially multiple calls to agents and tools) it will return a `Response` containing all the relevant updated state. Specifically, the new `messages`, the last `Agent` to be called, and the most up-to-date `context_variables`. You can pass these values (plus new user messages) in to your next execution of `client.run()` to continue the interaction where it left off – much like `chat.completions.create()`. (The `run_demo_loop` function implements an example of a full execution loop in `/swarm/repl/repl.py`.)

#### `Response` Fields

| Field                 | Type    | Description                                                                                                                                                                                                                                                                  |
| --------------------- | ------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Messages**          | `List`  | A list of message objects generated during the conversation. Very similar to [Chat Completions `messages`](https://platform.openai.com/docs/api-reference/chat/create#chat-create-messages), but with a `sender` field indicating which `Agent` the message originated from. |
| **Agent**             | `Agent` | The last agent to handle a message.                                                                                                                                                                                                                                          |
| **ContextVariables** | `dict`  | The same as the input variables, plus any changes.                                                                                                                                                                                                                           |

## Agents

An `Agent` simply encapsulates a set of `instructions` with a set of `functions` (plus some additional settings below), and has the capability to hand off execution to another `Agent`.

While it's tempting to personify an `Agent` as "someone who does X", it can also be used to represent a very specific workflow or step defined by a set of `instructions` and `functions` (e.g. a set of steps, a complex retrieval, single step of data transformation, etc). This allows `Agent`s to be composed into a network of "agents", "workflows", and "tasks", all represented by the same primitive.

## `Agent` Fields

| Field            | Type                     | Description                                                                   | Default                      |
| ---------------- | ------------------------ | ----------------------------------------------------------------------------- | ---------------------------- |
| **name**         | `str`                    | The name of the agent.                                                        | `"Agent"`                    |
| **option.WithAgentModel()**        | `str`                    | The model to be used by the agent.                                            | `"gpt-4o"`                   |
| **option.WithAgentInstructions()** | `str` or `func() -> str` | Instructions for the agent, can be a string or a callable returning a string. | `"You are a helpful agent."` |
| **option.WithAgentFunctions()**    | `List`                   | A list of functions that the agent can call.                                  | `[]`                         |
| **option.WithAgentToolChoice()**  | `str`                    | The tool choice for the agent, if any.                                        | `None`                       |

### Instructions

`Agent` `instructions` are directly converted into the `system` prompt of a conversation (as the first message). Only the `instructions` of the active `Agent` will be present at any given time (e.g. if there is an `Agent` handoff, the `system` prompt will change, but the chat history will not.)

```go
agent := goswarm.NewAgent(
   option.WithAgentInstructions("You are a helpful agent."),
)
```

The `instructions` can either be a regular `str`, or a function that returns a `str`. The function can optionally receive a `context_variables` parameter, which will be populated by the `context_variables` passed into `client.run()`.

```go

func Instructions(ctx goswarm.Context) string {
   userName := ctx.GetArg("user_name", "")
   return fmt.Sprintf("Help the user, %s, do whatever they want.", userName)
}

func main() {
	agent := goswarm.NewAgent(
	   option.WithAgentInstructions(Instructions),
	)

	contextVariables := types.Args{"user_name":"John"}

	messages := goswarm.NewMessages(openai.UserMessage("Hi!"))
	resp := client.Run(agent, messages, contextVariables)

	fmt.Println(resp.Messages[len(resp.Messages)-1].(openai.ChatCompletionMessage).Content)
```

```
Hi John, how can I assist you today?
```

## Functions

- Swarm `Agent`s can call python functions directly.
- Function should usually return a `str` (values will be attempted to be cast as a `str`).
- If a function returns an `Agent`, execution will be transferred to that `Agent`.
- If a function defines a `context_variables` parameter, it will be populated by the `context_variables` passed into `client.run()`.

```go

type GreetArgs struct {
	Language string `json:"language" desc:"language kind. e.g, [english, spanish]" required:"true"`
}

func Greet(ctx goswarm.Context, args GreetArgs) string {
   userName := ctx.GetArg("user_name", "")

   greeting := ""
   if args.Language == "spanish" {
     greeting = "Hola"
   } else {
     greeting = "Hello"
   }
   fmt.Printf("%s, %s!", greeting, userName)

   return "Done"
}

func main() {
	agent := goswarm.NewAgent("Agent"
	   option.WithAgentFunctions(Greet),
	)

	contextVariables := types.Args{"user_name":"John"}

	messages := goswarm.NewMessages(openai.UserMessage("Use greet() please."))
	resp := client.Run(agent, messages, contextVariables)

	fmt.Println(resp.Messages[len(resp.Messages)-1].(openai.ChatCompletionMessage).Content)
}
```

```
Hola, John!
```

- If an `Agent` function call has an error (missing function, wrong argument, error) an error response will be appended to the chat so the `Agent` can recover gracefully.
- If multiple functions are called by the `Agent`, they will be executed in that order.

### Handoffs and Updating Context Variables

An `Agent` can hand off to another `Agent` by returning it in a `function`.

```go

salesAgent := goswarm.NewAgent("Sales Agent")

transferToSales := func(ctx goswarm.Context) *types.Agent {
	return salesAgent
}

agent := goswarm.NewAgent("Agent",
   option.WithAgentFunctions(transferToSales),
)

messages := goswarm.NewMessages(openai.UserMessage("Transfer me to sales."))
resp := client.Run(agent, messages, types.Args{})

fmt.Println(resp.Agent.Name)

```

```
Sales Agent
```

It can also update the `context_variables` by returning a more complete `Result` object. This can also contain a `value` and an `agent`, in case you want a single function to return a value, update the agent, and update the context variables (or any subset of the three).

```go

salesAgent := goswarm.NewAgent("Sales Agent")

talkToSales := func(ctx goswarm.Context) *types.Agent {
	fmt.Println("Hello, World!")
	return &types.Result{
		Value: "Done",
		Agent: salesAgent,
		ContextVariables: types.Args{"department": "sales"},
	}
}

agent := goswarm.NewAgent("Agent",
   option.WithAgentFunctions(talkToSales),
)

messages := goswarm.NewMessages(openai.UserMessage("Transfer me to sales."))
response := client.Run(agent, messages, types.Args{"user_name": "John"})

fmt.Println(response.Agent.Name)
fmt.Println(response.ContextVariables)

```

```
Sales Agent
{'department': 'sales', 'user_name': 'John'}
```

> [!NOTE]
> If an `Agent` calls multiple functions to hand-off to an `Agent`, only the last handoff function will be used.

### Function Schemas

Swarm automatically converts functions into a JSON Schema that is passed into Chat Completions `tools`.

- Docstrings are turned into the function `description`.
- Parameters without default values are set to `required`.
- Type hints are mapped to the parameter's `type` (and default to `string`).
- Per-parameter descriptions are not explicitly supported, but should work similarly if just added in the docstring. (In the future docstring argument parsing may be added.)

```go

type GreetArgs struct {
	Name     string `json:"name"     desc:"Name of the user"     required:"true"`
	Age      int    `json:"age"      desc:"Age of the user"      required:"true"`
	Location string `json:"location" desc:"Best place on earth."`
}

func Greet(ctx goswarm.Context, args GreetArgs) string {
	if ctx.IsAnalyze() {
		ctx.SetDescription(
`Greets the user. Make sure to get their name and age before calling.

   Args:
      name: Name of the user.
      age: Age of the user.
      location: Best place on earth.`)
		return ""
	}

	if args.Location == "" {
		args.Location = "New York"
	}

   fmt.Printf("Hello %s, glad you are %d in %s!\n", args.Name, args.Age, args.Location)

   return ""
}
```

```javascript
{
   "type": "function",
   "function": {
      "name": "greet",
      "description": "Greets the user. Make sure to get their name and age before calling.\n\nArgs:\n   name: Name of the user.\n   age: Age of the user.\n   location: Best place on earth.",
      "parameters": {
         "type": "object",
         "properties": {
            "name": {"type": "string", "description": "Name of the user"},
            "age": {"type": "integer", "description": "Age of the user"},
            "location": {"type": "string", "description": "Best place on earth"}
         },
         "required": ["name", "age"]
      }
   }
}
```

## Streaming

```go
stream := client.RunAndStream(agent, messages, types.Args{}, option.WithStreamOption(true))
for chunk := range stream {
	print(chunk)
}
```

Uses the same events as [Chat Completions API streaming](https://platform.openai.com/docs/api-reference/streaming). See `ProcessAndPrintStreamingResponse` in `/go-swarm/repl/repl.go` as an example.

Two new event types have been added:

- `"start"` and `"end"`, to signal each time an `Agent` handles a single message (response or function call). This helps identify switches between `Agent`s.
- `Response` will return a `*types.Response` object at the end of a stream with the aggregated (complete) response, for convenience.

# Evaluations

Evaluations are crucial to any project, and we encourage developers to bring their own eval suites to test the performance of their swarms. For reference, we have some examples for how to eval swarm in the `airline`, `weather_agent` and `triage_agent` quickstart examples. See the READMEs for more details.

# Utils

Use the `RunDemoLoop` to test out your swarm! This will run a REPL on your command line. Supports streaming.

```go
import "github.com/chiwooi/go-swarm/repl"

repl.RunDemoLoop(agent, types.Args{}, option.WithStreamOption(true))
```

# Core Contributors

- chiwoo - [ibigio](https://github.com/chiwooi)



# reference

- [open-ai](https://github.com/openai/openai-go/tree/main/examples)
- [open-ai-swarm](https://github.com/openai/swarm)
