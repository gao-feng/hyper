package daemon

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/hyperhq/runv/hypervisor/types"
)

func (daemon Daemon) pausePod(podId string) error {
	daemon.PodList.RLock()
	glog.V(2).Infof("lock read of PodList")
	pod, ok := daemon.PodList.Get(podId)
	if !ok {
		glog.V(2).Infof("unlock read of PodList")
		daemon.PodList.RUnlock()
		return fmt.Errorf("Can not get Pod info with pod ID(%s)", podId)
	}
	vmId := pod.status.Vm
	glog.V(2).Infof("unlock read of PodList")
	daemon.PodList.RUnlock()

	vm, ok := daemon.VmList[vmId]
	if !ok {
		return fmt.Errorf("Can not find VM whose Id is %s!", vmId)
	}

	if err := vm.Pause(true); err != nil {
		return err
	}

	pod.status.SetContainerStatus(types.S_POD_PAUSED)
	pod.status.Status = types.S_POD_PAUSED
	vm.Status = types.S_VM_PAUSED

	return nil
}

func (daemon Daemon) PauseContainer(container string) error {
	glog.V(1).Infof("Get container id is %s", container)
	podId, err := daemon.GetPodByContainer(container)
	if err != nil {
		return err
	}

	return daemon.pausePod(podId)
}

func (daemon *Daemon) unpausePod(podId string) error {
	daemon.PodList.RLock()
	glog.V(2).Infof("lock read of PodList")
	pod, ok := daemon.PodList.Get(podId)
	if !ok {
		glog.V(2).Infof("unlock read of PodList")
		daemon.PodList.RUnlock()
		return fmt.Errorf("Can not get Pod info with pod ID(%s)", podId)
	}
	vmId := pod.status.Vm
	glog.V(2).Infof("unlock read of PodList")
	daemon.PodList.RUnlock()

	if pod.status.Status != types.S_POD_PAUSED {
		return fmt.Errorf("pod is not paused")
	}

	vm, ok := daemon.VmList[vmId]
	if !ok {
		return fmt.Errorf("Can not find VM whose Id is %s!", vmId)
	}

	if err := vm.Pause(false); err != nil {
		return err
	}

	pod.status.SetContainerStatus(types.S_POD_RUNNING)
	pod.status.Status = types.S_POD_RUNNING
	vm.Status = types.S_VM_ASSOCIATED

	return nil
}

func (daemon *Daemon) UnpauseContainer(container string) error {
	glog.V(1).Infof("Get container id is %s", container)
	podId, err := daemon.GetPodByContainer(container)
	if err != nil {
		return err
	}

	return daemon.unpausePod(podId)
}
