package pod

import (
	"io"

	"github.com/hyperhq/hyper/engine"
)

// Backend is the methods that need to be implemented to provide
// system specific functionality.
type Backend interface {
	CmdGetPodInfo(podName string) (*engine.Env, error)
	CmdGetPodStats(podId string) (*engine.Env, error)
	CmdCreatePod(podArgs string, autoremove bool) (*engine.Env, error)
	CmdSetPodLabels(podId string, override bool, labels map[string]string) (*engine.Env, error)
	CmdStartPod(in io.ReadCloser, out io.WriteCloser, podId, vmId, tag string) (*engine.Env, error)
	CmdList(item, podId, vmId string, auxiliary bool) (*engine.Env, error)
	CmdStopPod(podId, stopVm string) (*engine.Env, error)
	CmdCleanPod(podId string) (*engine.Env, error)
	CmdCreateVm(cpu, mem int, async bool) (*engine.Env, error)
	CmdKillVm(vmId string) (*engine.Env, error)
}
