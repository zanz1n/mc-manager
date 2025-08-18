package runner

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/zanz1n/mc-manager/config"
	"github.com/zanz1n/mc-manager/internal/pb"
)

type Runtime interface {
	Create(ctx context.Context, instance *Instance) error
	Launch(ctx context.Context, instance *Instance) error
	Stop(ctx context.Context, instance *Instance) error
}

type dockerRuntime struct {
	dockerPrefix    string
	dockerNetwork   string
	dockerNetworkId string
	dir             string

	docker *client.Client
	java   JavaVariant
	c      *http.Client
}

func NewDockerRuntime(
	ctx context.Context,
	cfg *config.Config,
	docker *client.Client,
	c *http.Client,
	java JavaVariant,
) (Runtime, error) {
	if c == nil {
		c = http.DefaultClient
	}

	dir, err := filepath.Abs(cfg.Data.DataDir)
	if err != nil {
		return nil, err
	}

	r := &dockerRuntime{
		dockerPrefix:  cfg.Docker.Prefix,
		dockerNetwork: cfg.Docker.NetworkName,
		dir:           dir,
		docker:        docker,
		java:          java,
		c:             c,
	}

	if err := r.createNetwork(ctx); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *dockerRuntime) createNetwork(ctx context.Context) error {
	nw, err := r.docker.NetworkInspect(ctx, r.dockerNetwork, network.InspectOptions{})
	if err != nil {
		nw, err := r.docker.NetworkCreate(ctx, r.dockerNetwork, network.CreateOptions{
			Driver: "bridge",
		})
		if err != nil {
			slog.Error(
				"DockerRunner: Failed to create network",
				"name", r.dockerNetwork,
				"error", err,
			)
			return err
		}
		slog.Info(
			"DockerRunner: Created network",
			"name", r.dockerNetwork,
			"id", nw.ID,
		)
		r.dockerNetworkId = nw.ID
	} else {
		slog.Info(
			"DockerRunner: Network fetched",
			"name", r.dockerNetwork,
			"id", nw.ID,
		)
		r.dockerNetworkId = nw.ID
	}

	return nil
}

func (r *dockerRuntime) Create(ctx context.Context, instance *Instance) error {
	dockerImage, err := r.pullImage(instance.Version.JavaVersion)
	if err != nil {
		return err
	}

	dataDir := path.Join(r.dir, instance.ID.String())
	if err = os.MkdirAll(dataDir, os.ModePerm); err != nil {
		return errors.Join(ErrFileSystem, err)
	}

	var hashb []byte
	if len(instance.Version.Hash) > 5 {
		hashb = instance.Version.Hash[0:4]
	} else {
		hashb = instance.Version.Hash
	}

	jarName := fmt.Sprintf("%s-%s-%s.jar",
		strings.ToLower(instance.Version.Distribution.String()),
		instance.Version.ID,
		hex.EncodeToString(hashb),
	)

	jarDir := path.Join(dataDir, jarName)

	if _, err = os.Stat(jarDir); err != nil {
		if os.IsNotExist(err) {
			err = instance.Version.DownloadTo(ctx, r.c, jarDir)
			if err != nil {
				return errors.Join(ErrInstanceCreate, err)
			}
		} else {
			return errors.Join(ErrFileSystem, err)
		}
	}

	if err = sanitizeMcProperties(dataDir, instance); err != nil {
		return errors.Join(ErrFileSystem, err)
	}

	if err = sanitizeEula(dataDir); err != nil {
		return errors.Join(ErrFileSystem, err)
	}

	containerName := r.dockerPrefix + "-" + instance.ID.String()
	portStr := strconv.Itoa(int(instance.Config.Port))

	cmd := makeJavaCommand(
		instance.Version.JVMArgs,
		jarName,
		instance.Limits.RAM,
	)

	res, err := r.docker.ContainerCreate(ctx,
		&container.Config{
			Image:        dockerImage,
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
			OpenStdin:    true,
			Tty:          true,
			// WorkingDir:   "/home/container",
			WorkingDir: "/game",
			// User:       "container",
			// Env:        []string{"USER=container", "HOME=/home/container"},
			Cmd: cmd,
		},
		&container.HostConfig{
			AutoRemove: true,
			Resources: container.Resources{
				CPUPercent: int64(instance.Limits.CPU),
				Memory:     int64(instance.Limits.RAM),
			},
			PortBindings: nat.PortMap{
				nat.Port(portStr): []nat.PortBinding{
					nat.PortBinding{
						HostIP:   "0.0.0.0",
						HostPort: portStr,
					},
				},
			},
			Mounts: []mount.Mount{{
				Type:   mount.TypeBind,
				Source: dataDir,
				Target: "/game",
			}},
		},
		&network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				r.dockerNetwork: &network.EndpointSettings{
					NetworkID: r.dockerNetworkId,
				},
			},
		},
		nil,
		containerName,
	)
	if err != nil {
		return errors.Join(ErrInstanceCreate, err)
	}

	instance.ContainerID = res.ID
	return nil
}

func (r *dockerRuntime) Launch(ctx context.Context, instance *Instance) error {
	if instance.ContainerID == "" {
		return errors.Join(
			ErrInstanceLaunch,
			errors.New("instance not created yet"),
		)
	}

	err := r.docker.ContainerStart(ctx, instance.ContainerID, container.StartOptions{})
	if err != nil {
		return errors.Join(ErrInstanceLaunch, err)
	}

	res, err := r.docker.ContainerAttach(ctx, instance.ContainerID, container.AttachOptions{
		Stream: true,
		Stdin:  true,
		Stdout: true,
		Stderr: true,
	})
	if err != nil {
		return errors.Join(ErrInstanceLaunch, err)
	}

	instance.setStream(res)
	instance.SetState(pb.InstanceState_STATE_STARTING)
	instance.launch()

	return nil
}

func (r *dockerRuntime) Stop(ctx context.Context, instance *Instance) error {
	if instance.ContainerID == "" {
		return errors.Join(
			ErrInstanceStop,
			errors.New("instance not created yet"),
		)
	}

	if !instance.Launched.Load() {
		return errors.Join(
			ErrInstanceStop,
			errors.New("instance not launched yet"),
		)
	}

	instance.SendEvent(Event{
		Type: pb.EventType_EVENT_SHUTTING_DOWN,
	})
	instance.SetState(pb.InstanceState_STATE_SHUTTING_DOWN)

	timeout := 20

	err := r.docker.ContainerStop(ctx, instance.ContainerID, container.StopOptions{
		Timeout: &timeout,
		Signal:  "SIGINT",
	})
	if err != nil {
		return errors.Join(ErrInstanceStop, err)
	}

	instance.SetState(pb.InstanceState_STATE_OFFLINE)
	instance.SendEvent(Event{
		Type: pb.EventType_EVENT_STOPPED,
	})
	instance.close()

	return nil
}

func (r *dockerRuntime) pullImage(v pb.JavaVersion) (string, error) {
	ref, err := r.java.GetImage(v)
	if err != nil {
		return "", err
	}

	res, err := r.docker.ImagePull(context.Background(), ref, image.PullOptions{})
	if err != nil {
		return "", errors.Join(ErrJavaVersion, err)
	}
	defer res.Close()

	b := make([]byte, 1024)
	for {
		_, err = res.Read(b)
		if err != nil {
			break
		}
	}

	return ref, nil
}
