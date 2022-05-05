package v1

type ResourceList map[string]string
type NodePhase string

type Node struct {
	TypeMeta

	ObjectMeta

	Status NodeStatus `json:"status,omitempty"`
}

type NodeStatus struct {
	// some capacity, such as disk size, cpu usage...
	// what's needed by scheduler?
	Capacity ResourceList

	Available ResourceList

	Phase NodePhase

	Images []ContainerImage

	VolumesAttached []AttachedVolume
}

type ContainerImage struct {
	Name string `json:"name"`

	SizeBytes int64 `json:"sizeBytes"`
}

type AttachedVolume struct {
	Name string `json:"name"`

	DevicePath string `json:"devicePath"`
}
