package goswarm

import (
	"context"
	"time"

    "github.com/chiwooi/go-swarm/types"
)

type Context interface {
	Deadline() (deadline time.Time, ok bool)
	Done() <-chan struct{}
	Err() error
	Value(key any) any

	GetVariables() types.ContextVariables
	SetVariables(vars types.ContextVariables)
	GetVariable(name string, def any) any
	SetVariable(name string, val any)

	SetAnalyze(flag bool)
	IsAnalyze() bool
	SetDescription(desc string)
	GetDescription() string
	GetContext() context.Context
}

type argsContext struct {
	context.Context
}

func (c argsContext) Deadline() (deadline time.Time, ok bool) {
	return c.Context.Deadline()
}

func (c argsContext) Done() <-chan struct{} {
	return c.Context.Done()
}

func (c argsContext) Err() error {
	return c.Context.Err()
}

func (c argsContext) Value(key interface{}) any {
	return c.Context.Value(key)
}


type variablesKey struct{}

func (c *argsContext) GetVariables() types.ContextVariables {
	if v := c.Context.Value(variablesKey{}); v != nil {
		return v.(types.ContextVariables)
	}

	vars := types.ContextVariables{}
	c.SetVariables(vars)

	return vars
}

func (c *argsContext) SetVariables(vars types.ContextVariables) {
	c.Context = context.WithValue(c.Context, variablesKey{}, vars)
}

func (c *argsContext) GetVariable(name string, def any) any {
	if vars := c.GetVariables(); vars != nil {
		return vars.Get(name, def)
	}
	return nil
}

func (c *argsContext) SetVariable(name string, val any) {
	if vars := c.GetVariables(); vars != nil {
		vars.Set(name, val)
	}
}

// This is a flag to indicate if the function is being called for analysis purposes.
type analyzeKey struct{}

func (c argsContext) IsAnalyze() bool {
	if flag := c.Context.Value(analyzeKey{}); flag != nil {
		return flag.(bool)
	}
	return false
}

func (c *argsContext) SetAnalyze(flag bool)  {
	c.Context = context.WithValue(c.Context, analyzeKey{}, flag)
	return
}

type fnDescKey struct{}

func (c argsContext) GetDescription() string {
	if desc := c.Context.Value(fnDescKey{}); desc != nil {
		return desc.(string)
	}
	return ""
}

func (c *argsContext) SetDescription(desc string) {
	c.Context = context.WithValue(c.Context, fnDescKey{}, desc)
	return
}

func (c argsContext) GetContext() context.Context {
	return c.Context
}

func NewContext(parent context.Context) Context {
	return &argsContext{parent}
}


