package option

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
   Model:        "gpt-4o",
   MaxTurns:     9999,
   ExecuteTools: true,
   Stream:       false,
   Debug:        false,
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

func WithStream(flag bool) StreamOption {
   return StreamOption(flag)
}


type DebugOption bool

func (o DebugOption) ApplyOption(opts *RunOptions) {
   opts.Debug = bool(o)
}

func WithDebug(flag bool) DebugOption {
   return DebugOption(flag)
}


type MaxTurnsOption int

func (o MaxTurnsOption) ApplyOption(opts *RunOptions) {
   opts.MaxTurns = int(o)
}

func WithMaxTurns(maxTurns int) MaxTurnsOption {
   return MaxTurnsOption(maxTurns)
}


type ExecuteToolsOption bool

func (o ExecuteToolsOption) ApplyOption(opts *RunOptions) {
   opts.ExecuteTools = bool(o)
}

func WithExecuteTools(exec bool) ExecuteToolsOption {
   return ExecuteToolsOption(exec)
}
