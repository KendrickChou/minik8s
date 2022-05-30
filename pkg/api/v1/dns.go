package v1

type DNS struct {
	TypeMeta

	ObjectMeta `json:"metadata,omitempty"`

	Host string `json:"host,omitempty"`

	Paths []DNSPath `json:"paths,omitempty"`
}

type DNSPath struct {
	Path string `json:"path,omitempty"`

	ServiceName string `json:"servicename,omitempty"`

	Port int `json:"port,omitempty"`
}
