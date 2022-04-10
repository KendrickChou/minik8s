package container

import (
	"context"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"k8s.io/klog/v2"
)

type Manager interface {
	CreateContainer(config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, containerName string) (string, error)
	StartContainer(containerID string, options types.ContainerStartOptions) error
	StopContainer(containerID string, timeout time.Duration) error
	RemoveContainer(containerID string, options types.ContainerRemoveOptions) error
	ListContainers(options types.ContainerListOptions) ([]types.Container, error)
	ContainerStatus(containerID string)
	UpdateContainerResources()	
}

var Platform v1.Platform = v1.Platform{
								Architecture: "x86_64",
								OS: "linux",}

type DockerContainerManager struct {
	cli *client.Client
}

func (manager *DockerContainerManager)Init() {
	cli, err := client.NewClientWithOpts(client.WithVersion("v20.10.14+incompatible"))
	manager.cli = cli

	if err != nil {
		klog.Fatalf("Create Docker client failed")
	}
}

func (manager *DockerContainerManager) CreateContainer(config *container.Config,
	hostConfig *container.HostConfig,
	networkingConfig *network.NetworkingConfig,
	containerName string) (string, error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	containerBody, err := manager.cli.ContainerCreate(ctx, config, hostConfig, networkingConfig, &Platform, containerName)
	defer cancel()

	if err != nil {
		klog.Errorf("Create container %s failed", containerName)
		return "", err
	}

	if len(containerBody.Warnings) != 0 {
		for warning := range containerBody.Warnings {
			klog.Infof("Create container warning %s", warning)
		}
	}

	return containerBody.ID, nil
}

func (manager *DockerContainerManager) StartContainer(containerID string, options types.ContainerStartOptions) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	err := manager.cli.ContainerStart(ctx, containerID, options)
	defer cancel()

	if err != nil {
		klog.Errorln("Start container " + containerID + " failed:", err)
	}

	return err
}

func (manager *DockerContainerManager) StopContainer(containerID string, timeout time.Duration) error {
	ctx := context.Background()
	err := manager.cli.ContainerStop(ctx, containerID, &timeout)

	if err != nil {
		klog.Errorln(err)
	}

	return err
}

func (manager *DockerContainerManager) RemoveContainer(containerID string, options types.ContainerRemoveOptions) error {
	ctx := context.Background()
	err := manager.cli.ContainerRemove(ctx, containerID, options)

	if err != nil {
		klog.Errorln(err)
	}

	return err
}

func (manager *DockerContainerManager) ListContainers(options types.ContainerListOptions) ([]types.Container, error) {
	ctx := context.Background()
	containers, err := manager.cli.ContainerList(ctx, options)

	if err != nil {
		klog.Errorln(err)
		return nil, err
	}

	return containers, nil
}

func (manager *DockerContainerManager) ContainerStatus(containerID string) {
	//TODO
}





