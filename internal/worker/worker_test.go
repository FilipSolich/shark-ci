package worker

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	dockertypes "github.com/docker/docker/api/types"
	containertypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

//func TestProcessWork(t *testing.T) {
//	objStore := objectstore.NewMocObjectStore()
//	processWork(context.TODO(), objStore, types.Work{})
//}

func TestDockerStart(t *testing.T) {
	t.Skip()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	defer cli.Close()
	container, err := cli.ContainerCreate(context.Background(), &containertypes.Config{
		Image: "ubuntu",
		Tty:   true,
	}, &containertypes.HostConfig{Binds: []string{"/tmp/test:/app"}}, nil, nil, "MyContainer")
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	err = cli.ContainerStart(context.Background(), container.ID, containertypes.StartOptions{})
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	exec, err := cli.ContainerExecCreate(context.Background(), container.ID, dockertypes.ExecConfig{
		AttachStdout: true,
		AttachStderr: true,
		Detach:       true,
		Cmd:          []string{"echo", "hello me"},
	})
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	hijacked, err := cli.ContainerExecAttach(context.Background(), exec.ID, dockertypes.ExecStartCheck{})
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	logsBuff := &bytes.Buffer{}
	_, err = stdcopy.StdCopy(logsBuff, logsBuff, hijacked.Reader)
	if err != nil {
		hijacked.Close()
		t.Errorf("Error: %v", err)
	}
	hijacked.Close()
	fmt.Println(logsBuff.String())

	err = cli.ContainerKill(context.Background(), container.ID, "SIGKILL")
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	// Delete container.
	err = cli.ContainerRemove(context.Background(), container.ID, containertypes.RemoveOptions{Force: true})
	if err != nil {
		t.Errorf("Error: %v", err)
	}
}
