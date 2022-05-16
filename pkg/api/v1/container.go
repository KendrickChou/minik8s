package v1

// ContainerState holds a possible state of container.
// Only one of its members may be specified.
// If none of them is specified, the default one is ContainerStateWaiting.
type ContainerState struct {
	// String representation of the container state. Can be one of "created", "running", "paused", "restarting", "removing", "exited", or "dead"
	Status string `json:"status"`
	// Exit status from the last termination of the container
	ExitCode int `json:"exitCode"`
	// Error info
	Error string `json:"reason,omitempty"`
	// Time at which previous execution of the container started
	StartedAt string `json:"startedAt,omitempty"`
	// Time at which the container last terminated
	FinishedAt string `json:"finishedAt,omitempty"`
}

type Container struct {
	Name string `json:"Name,omitempty"`

	Namespace string `json:"Namespace,omitempty"`

	ID string `json:"ID,omitempty"`

	Image string `json:"Image,omitempty"`

	//"Always" means that kubelet always attempts to pull the latest image. Container will fail If the pull fails.
	//"IfNotPresent" means that kubelet pulls if the image isn't present on disk. Container will fail if the image isn't present and the pull fails.
	//"Never" means that kubelet never pulls an image, but only uses a local image. Container will fail if the image isn't present
	//default: IfNotPresent
	ImagePullPolicy string `json:"ImagePullPolicy,omitempty" default:"IfNotPresent"`

	// Command to run when starting the container
	Command []string `json:"Command,omitempty"`

	Entrypoint []string `json:"Entrypoint,omitempty"`

	// Container's working directory.
	// If not specified, the container runtime's default will be used, which
	// might be configured in the container image.
	WorkingDir string `json:"WorkingDir,omitempty"`

	Env []string `json:"Env,omitempty"`

	// mount volumes
	Mounts []Mount `json:"Mounts,omitempty"`

	NetworkMode string `json:"-"`

	DNS string `json:"-"`

	DNSSearch string `json:"-"`
}

type ContainerStatus struct {
	Name string `json:"Name,omitempty"`

	State ContainerState `json:"State,omitempty"`
}

type Mount struct {
	Type        string `json:"MountType,omitempty"`
	Source      string `json:"MountSource,omitempty"`
	Target      string `json:"MountTarget,omitempty"`
	Consistency string `json:"MountConsistency,omitempty"`
}

const (
	AlwaysImagePullPolicy       string = "Always"
	IfNotPresentImagePullPolicy string = "IfNotPresent"
	NeverPullPolicy             string = "Never"

	// TypeBind is the type for mounting host dir
	TypeBind string = "bind"
	// TypeVolume is the type for remote storage volumes
	TypeVolume string = "volume"

	// ConsistencyFull guarantees bind mount-like consistency
	ConsistencyFull string = "consistent"
	// ConsistencyCached mounts can cache read data and FS structure
	ConsistencyCached string = "cached"
	// ConsistencyDelegated mounts can cache read and written data and structure
	ConsistencyDelegated string = "delegated"
	// ConsistencyDefault provides "consistent" behavior unless overridden
	ConsistencyDefault string = "default"
)
