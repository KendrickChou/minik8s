package v1

type GPUJob struct {
	TypeMeta

	ObjectMeta `json:"metadata,omitempty"`

	Files []UploadFile `json:"files,omitempty"`

	Script string `json:"script,omitempty"`

	JobNum string `json:"jobnum,omitempty"`

	Output string `json:"output,omitempty"`

	Error string `json:"error,omitempty"`
}

type UploadFile struct {
	Filename string `json:"filename,omitempty"`
}
