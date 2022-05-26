package actionchain

const (
	ACT_TASK   = 0
	ACT_CHOICE = 1
)

type ActionType int

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

	// int64 / bool / string
	VarType string `json:"VarType"`

	NumericEqual int64 `json:"NumericEqual"`

	BooleanEqual bool `json:"BooleanEqual"`

	StringEqual string `json:"StringEqual"`

	Next string `json:"Next"`
}
