package v1

type Node struct {
	TypeMeta

	ObjectMeta

	Status NodeStatus `json:"status,omitempty"`
}

type NodeStatus struct {
	Phase string `json:"phase,omitempty"`
}
