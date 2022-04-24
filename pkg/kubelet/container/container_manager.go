package container

import (
	"context"
	"errors"
	"io"
	"os"
	"time"

	"github.com/docker/docker/api/types"
	dockerctnr "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"k8s.io/klog/v2"

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
	ContainerStatus(ctx context.Context, containerID string) (*types.ContainerState, error)
}

type containerManager struct {
	dockerClient *client.Client
}

const (
	restartTimeout time.Duration = time.Duration(time.Second * 3)
)

func NewContainerManager() (ContainerManager, error) {

	cli, err := client.NewClientWithOpts(client.FromEnv)

	cm := &containerManager{dockerClient: cli}

	return cm, err
}

func (manager *containerManager) CreateContainer(ctx context.Context, container *v1.Container) (string, error) {
	klog.Infof("Create Container %v", container.Name)

	containerConfig := &dockerctnr.Config{}
	hostConfig := &dockerctnr.HostConfig{}

	switch {
	case container.Image != "":
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
			return "", errors.New("Unknown Image Pull Policy")
		}

		fallthrough
	case len(container.Command) != 0:
		containerConfig.Cmd = append(containerConfig.Cmd, container.Command...)
		fallthrough
	case len(container.Entrypoint) != 0:
		containerConfig.Entrypoint = append(containerConfig.Entrypoint, container.Entrypoint...)
		fallthrough
	case len(container.Env) != 0:
		containerConfig.Env = append(containerConfig.Env, container.Env...)
		fallthrough
	case len(container.Mounts) != 0:
		for _, m := range container.Mounts {
			hostConfig.Mounts = append(hostConfig.Mounts, mount.Mount{
				Type:        mount.Type(m.Type),
				Source:      m.Source,
				Target:      m.Target,
				Consistency: mount.Consistency(m.Consistency),
			})
		}
		fallthrough
	default:
	}

	body, err := manager.dockerClient.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, container.Name)

	return body.ID, err
}

func (manager *containerManager) StartContainer(ctx context.Context, container *v1.Container) error {
	err := manager.dockerClient.ContainerStart(ctx, container.ID, types.ContainerStartOptions{})

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

func (manager *containerManager) ContainerStatus(ctx context.Context, containerID string) (*types.ContainerState, error) {
	cntr, err := manager.dockerClient.ContainerInspect(ctx, containerID)

	return cntr.State, err
}
