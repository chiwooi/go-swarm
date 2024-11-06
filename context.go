package goswarm

import (
	"context"

    "github.com/chiwooi/go-swarm/types"
)

type Context interface {
	Args() *types.Args
	GetArg(name string, def any) any
	IsAnalyze() bool
	SetDescription(desc string)
	GetDescription() string
	GetContext() context.Context
}

type argsContext struct {
	context.Context
	args    *types.Args
	analyze bool  // This is a flag to indicate if the function is being called for analysis purposes.
	fnDesc  string
}

func (c argsContext) Args() *types.Args {
	return c.args
}

func (c *argsContext) GetArg(name string, def any) any {
	return c.args.Get(name, def)
}

func (c argsContext) IsAnalyze() bool {
	return c.analyze
}

func (c *argsContext) SetDescription(desc string) {
	c.fnDesc = desc
}

func (c argsContext) GetDescription() string {
	return c.fnDesc
}

func (c argsContext) GetContext() context.Context {
	return c.Context
}

func AnalyzeContext(args types.Args) Context {
	return &argsContext{
		Context: context.Background(),
		analyze: true,
		args:    &args,
	}
}

func RunContext(args types.Args) Context {
	return &argsContext{
		Context: context.Background(),
		analyze: false,
		args:    &args,
	}
}
