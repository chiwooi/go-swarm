package goswarm

type RunOption interface {
   ApplyOption(opts *RunOptions)
}

type RunOptions struct {
	Model         string
	Stream        bool
	Debug         bool
	MaxTurns      int
	ExecuteTools  bool
}

var DefRunOptions = RunOptions{
   Model:    "",
   MaxTurns: 9999,
   Debug:    true,
}

type ModelOption string

func (o ModelOption) ApplyOption(opts *RunOptions) {
   opts.Model = string(o)
}

func WithModel(model string) ModelOption {
   return ModelOption(model)
}


type StreamOption bool

func (o StreamOption) ApplyOption(opts *RunOptions) {
   opts.Stream = bool(o)
}

func WithStreamOption(flag bool) StreamOption {
   return StreamOption(flag)
}


type DebugOption bool

func (o DebugOption) ApplyOption(opts *RunOptions) {
   opts.Debug = bool(o)
}

func WithDebugOption(flag bool) DebugOption {
   return DebugOption(flag)
}


type MaxTurnsOption int

func (o MaxTurnsOption) ApplyOption(opts *RunOptions) {
   opts.MaxTurns = int(o)
}

func WithMaxTurnsOption(maxTurns int) MaxTurnsOption {
   return MaxTurnsOption(maxTurns)
}


type ExecuteToolsOption bool

func (o ExecuteToolsOption) ApplyOption(opts *RunOptions) {
   opts.ExecuteTools = bool(o)
}

func WithExecuteToolsOption(exec bool) ExecuteToolsOption {
   return ExecuteToolsOption(exec)
}
