package container

import (
	"io"

	"github.com/docker/engine-api/types"
	"github.com/hyperhq/hyper/daemon"
	"github.com/hyperhq/hyper/engine"
)

type Backend interface {
	CmdGetContainerInfo(container string) (*engine.Env, error)
	CmdGetContainerLogs(name string, c *daemon.ContainerLogsConfig) error
	CmdExitCode(container, tag string) (int, error)
	CmdCreateContainer(types.ContainerCreateConfig) (*engine.Env, error)
	CmdContainerRename(oldName, newName string) (*engine.Env, error)
	CmdExec(in io.ReadCloser, out io.WriteCloser, key, id, cmd, tag string) error
	CmdAttach(in io.ReadCloser, out io.WriteCloser, key, id, tag string) error
	CmdCommitImage(name string, cfg *types.ContainerCommitConfig) (*engine.Env, error)
	CmdTtyResize(podId, tag string, h, w int) error
}
