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

	CPUPerc string `json:"cpuperc,omitempty"` // cpu usage ratio, e.g. 0.01%

	// MemUsage string `json:"memusage,omitempty"` // mem usage, MEM USAGE / LIMIT, e.g. 4MiB / 7GiB

	MemPerc string `json:"memperc,omitempty"` // mem usage ratio, e.g. 1%

	// NetworkIO string `json:"networkio,omitempty"` // networkio, e.g. 10.6kB / 18.1kB
}

type Container struct {
	Name string `json:"name,omitempty"`

	Namespace string `json:"namespace,omitempty"`

	ID string `json:"id,omitempty"`

	Image string `json:"image,omitempty"`

	//"Always" means that kubelet always attempts to pull the latest image. Container will fail If the pull fails.
	//"IfNotPresent" means that kubelet pulls if the image isn't present on disk. Container will fail if the image isn't present and the pull fails.
	//"Never" means that kubelet never pulls an image, but only uses a local image. Container will fail if the image isn't present
	//default: IfNotPresent
	ImagePullPolicy string `json:"imagepullpolicy,omitempty" default:"IfNotPresent"`

	// Command to run when starting the container
	Command []string `json:"command,omitempty"`

	Entrypoint []string `json:"entrypoint,omitempty"`

	// Container's working directory.
	// If not specified, the container runtime's default will be used, which
	// might be configured in the container image.
	WorkingDir string `json:"workingdir,omitempty"`

	Env []string `json:"env,omitempty"`

	// mount volumes
	Mounts []Mount `json:"mounts,omitempty"`

	NetworkMode string `json:"-"`

	DNS string `json:"-"`
 
	DNSSearch string `json:"-"`

	ExposedPorts []string `json:"-"`

	BindPorts map[string]string `json:"-"`

	// --- resource type --- | --- value ---
	//          cpu					1
	//			memory				128 * 1024 * 1024 in bytes
	Resources map[string]string `json:"resources,omitempty"`
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
