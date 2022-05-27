package actionchain

const (
	ACT_TASK   = 0
	ACT_CHOICE = 1

	VAR_INT    = 0
	VAR_STRING = 1
	VAR_BOOL   = 2
)

type ActionType int
type VarType int

type ActionChain struct {
	StartAt string `json:"StartAt"`

	Chain map[string]Action `json:"Chain"`
}

type Action struct {
	// Task / Choice
	Type ActionType `json:"Type"`

	// e.g. GO, Python, C++...
	Env string `json:"Env"`

	// corresponding function id
	Function string `json:"Function"`

	// use for Task
	Next string `json:"Next"`

	// use for choice
	Choices []Choice `json:"Choices,omitempty"`

	// default is false
	End bool `json:"End"`
}

type Choice struct {
	Variable string `json:"Variable"`

	Type VarType `json:"Type"`

	NumericEqual int64 `json:"NumericEqual"`

	BooleanEqual bool `json:"BooleanEqual"`

	StringEqual string `json:"StringEqual"`

	Next string `json:"Next"`
}
