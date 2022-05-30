package container

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types"
	dockerctnr "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"k8s.io/klog/v2"
	bytesize "github.com/inhies/go-bytesize"
	v1 "minik8s.com/minik8s/pkg/api/v1"
)

type ContainerManager interface {
	CreateContainer(ctx context.Context, container *v1.Container) (string, error)
	StartContainer(ctx context.Context, container *v1.Container) error
	RestartContainer(ctx context.Context, container *v1.Container) error
	StopContainer(ctx context.Context, container *v1.Container) error
	PauseContainer(ctx context.Context, container *v1.Container) error
	ResumeContainer(ctx context.Context, container *v1.Container) error
	RemoveContainer(ctx context.Context, container *v1.Container) error
	ListContainers(ctx context.Context) ([]*v1.Container, error)
	ContainerStatus(ctx context.Context, containerID string) (types.ContainerJSON, error) // some static status, e.g. IP, Networkmode
	CreateNetwork(ctx context.Context, name string, CIDR string) (string, error)
	RemoveNetwork(ctx context.Context, networkID string) error
	ConnectNetwork(ctx context.Context, networkID string, containerID string) error
	ListNetwork(ctx context.Context, filter types.NetworkListOptions) ([]types.NetworkResource, error)
	ContainerStats(ctx context.Context, containerID string) ([]string, error) // some dynamic status, e.g. cpu, mem usage, net
}

type containerManager struct {
	dockerClient *client.Client
}

func NewContainerManager() (ContainerManager, error) {

	cli, err := client.NewClientWithOpts(client.FromEnv)

	cm := &containerManager{dockerClient: cli}

	return cm, err
}

func (manager *containerManager) CreateContainer(ctx context.Context, container *v1.Container) (string, error) {
	klog.Infof("Create Container %v", container.Name)

	containerConfig := &dockerctnr.Config{}
	hostConfig := &dockerctnr.HostConfig{}

	if container.Image != "" {
		containerConfig.Image = container.Image

		switch container.ImagePullPolicy {
		case v1.AlwaysImagePullPolicy:
			out, err := manager.dockerClient.ImagePull(ctx, containerConfig.Image, types.ImagePullOptions{})

			if err != nil {
				return "", err
			}

			defer out.Close()
			io.Copy(os.Stdout, out)

		case v1.IfNotPresentImagePullPolicy:
			var isExist bool = false
			imageList, _ := manager.dockerClient.ImageList(ctx, types.ImageListOptions{})
			for _, i := range imageList {
				for _, tag := range i.RepoTags {
					if tag == containerConfig.Image {
						isExist = true
						break
					}
				}

				if isExist {
					break
				}
			}
			if !isExist {
				out, err := manager.dockerClient.ImagePull(ctx, containerConfig.Image, types.ImagePullOptions{})

				if err != nil {
					return "", err
				}

				defer out.Close()

				io.Copy(os.Stdout, out)
			}
		case v1.NeverPullPolicy:
			// do nothing
		default:
			return "", errors.New("unknown image Pull Policy")
		}
	}
	if container.NetworkMode != "" {
		hostConfig.NetworkMode = dockerctnr.NetworkMode(container.NetworkMode)
	}
	if container.DNS != "" {
		hostConfig.DNS = []string{container.DNS}
	}
	if container.DNSSearch != "" {
		hostConfig.DNSSearch = []string{container.DNSSearch}
	}
	if len(container.Command) != 0 {
		containerConfig.Cmd = append(containerConfig.Cmd, container.Command...)
	}
	if len(container.Entrypoint) != 0 {
		containerConfig.Entrypoint = append(containerConfig.Entrypoint, container.Entrypoint...)
	}
	if len(container.Env) != 0 {
		containerConfig.Env = append(containerConfig.Env, container.Env...)
	}
	if len(container.Mounts) != 0 {
		for _, m := range container.Mounts {
			hostConfig.Mounts = append(hostConfig.Mounts, mount.Mount{
				Type:        mount.Type(m.Type),
				Source:      m.Source,
				Target:      m.Target,
				Consistency: mount.Consistency(m.Consistency),
			})
		}
	}
	if len(container.ExposedPorts) != 0 {
		exposedPorts := nat.PortSet{}

		for _, port := range container.ExposedPorts {
			exposedPorts[nat.Port(port)] = struct{}{}
		}

		containerConfig.ExposedPorts = exposedPorts
	}
	if len(container.BindPorts) != 0 {
		portMap := nat.PortMap{}

		for port, hostaddr := range container.BindPorts {
			array := strings.Split(hostaddr, ":")
			if len(array) != 2 {
				klog.Error("Parsed BindPorts Error ", array)
				continue
			}
			host := array[0]
			hostport := array[1]
			portMap[nat.Port(port)] = []nat.PortBinding{
				{
					HostIP: host,
					HostPort: hostport,
				},
			}
		}

		hostConfig.PortBindings = portMap
	}
	if len(container.Resources) != 0 {
		resources := dockerctnr.Resources{}
		cpu, ok := container.Resources["cpu"]

		if ok {
			num, _ := strconv.ParseInt(cpu, 10, 64)
			resources.NanoCPUs = 1000000000 * num
		}
		
		mem, ok := container.Resources["memory"]

		if ok {
			inBytes, err := bytesize.Parse(mem)

			if err != nil {
				klog.Errorf("Container %s Parse Memory Size Error: %s", container.Name, inBytes)
			} else {
				resources.Memory = int64(inBytes)
			}
		}

		hostConfig.Resources = resources
	}

	body, err := manager.dockerClient.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, container.Name)

	return body.ID, err
}

func (manager *containerManager) StartContainer(ctx context.Context, container *v1.Container) error {
	klog.Infof("Start Container %v", container.Name)
	manager.dockerClient.ContainerStart(ctx, container.ID, types.ContainerStartOptions{})

	out, err := manager.dockerClient.ContainerLogs(ctx, container.ID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true})
	defer out.Close()

	io.Copy(os.Stdout, out)

	return err
}

func (manager *containerManager) RestartContainer(ctx context.Context, container *v1.Container) error {
	err := manager.dockerClient.ContainerRestart(ctx, container.ID, nil)
	return err
}

func (manager *containerManager) StopContainer(ctx context.Context, container *v1.Container) error {
	klog.Infof("Stop Container %s", container.Name)
	err := manager.dockerClient.ContainerStop(ctx, container.ID, nil)
	return err
}

func (manager *containerManager) PauseContainer(ctx context.Context, container *v1.Container) error {
	err := manager.dockerClient.ContainerPause(ctx, container.ID)
	return err
}

func (manager *containerManager) ResumeContainer(ctx context.Context, container *v1.Container) error {
	err := manager.dockerClient.ContainerUnpause(ctx, container.ID)
	return err
}

func (manager *containerManager) RemoveContainer(ctx context.Context, container *v1.Container) error {
	klog.Infof("Remove Container %s", container.Name)

	err := manager.dockerClient.ContainerRemove(ctx, container.ID, types.ContainerRemoveOptions{})
	return err
}

func (manager *containerManager) ListContainers(ctx context.Context) ([]*v1.Container, error) {
	// do we really need this?

	return nil, nil
}

func (manager *containerManager) ContainerStatus(ctx context.Context, containerID string) (types.ContainerJSON, error) {
	cntr, err := manager.dockerClient.ContainerInspect(ctx, containerID)

	if err != nil {
		return types.ContainerJSON{}, err
	}

	return cntr, err
}

func (manager *containerManager) CreateNetwork(ctx context.Context, name string, CIDR string) (string, error) {
	resp, err := manager.dockerClient.NetworkCreate(ctx, name, types.NetworkCreate{
		CheckDuplicate: true,
		Driver:         "bridge",
		IPAM: &network.IPAM{
			Driver: "default",
			Config: []network.IPAMConfig{
				{
					Subnet: CIDR,
					// Gateway: constants.GatewayAddress,
				},
			},
		},
	})

	if err != nil {
		return "", err
	}

	return resp.ID, nil
}

func (manager *containerManager) RemoveNetwork(ctx context.Context, networkID string) error {
	err := manager.dockerClient.NetworkRemove(ctx, networkID)
	return err
}

func (manager *containerManager) ConnectNetwork(ctx context.Context, networkID string, containerID string) error {
	klog.Infof("Connect to Network: %s", networkID)

	err := manager.dockerClient.NetworkConnect(ctx, networkID, containerID, nil)
	return err
}

func (manager *containerManager) ListNetwork(ctx context.Context, filter types.NetworkListOptions) ([]types.NetworkResource, error) {
	networks, err := manager.dockerClient.NetworkList(ctx, filter)

	if err != nil {
		return nil, err
	}

	return networks, nil
}

func (manager *containerManager) ContainerStats(ctx context.Context, containerID string) ([]string, error) {
	command := fmt.Sprintf("docker stats %s --no-stream --format \"{{.CPUPerc}}\t{{.MemPerc}}\"", containerID)
	cmd := exec.Command("bash", "-c", command)
	res, err := cmd.CombinedOutput()

	if err != nil {
		klog.Errorf("Get Container Stats Error: %s", err.Error())
		return nil, err
	}

	return strings.Split(string(res[:(len(res)-1)]), "\t"), nil
}
