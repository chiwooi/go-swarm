# [Triage agent](https://github.com/openai/swarm/tree/main/examples/triage_agent)

This example is a Swarm containing a triage agent, which takes in user inputs and chooses whether to respond directly, or triage the request
to a sales or refunds agent.

## Setup

To run the triage agent Swarm:

1. Run

```shell
go run .
```

## Evals

> [!NOTE]
> These evals are intended to be examples to demonstrate functionality, but will have to be updated and catered to your particular use case.

This example uses `testing` to run eval unit tests. We have two tests in the `evals_test.go` file, one which
tests if we call the correct triage function when expected, and one which assesses if a conversation
is 'successful', as defined in our prompt in `evals_test.go`.

To run the evals, run

```shell
go test agents.go evals_util.go evals_test.go -v
```
