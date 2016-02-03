package pod

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/golang/glog"
	"github.com/hyperhq/hyper/server/httputils"
	"golang.org/x/net/context"
)

func (p *podRouter) getPodInfo(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if err := httputils.ParseForm(r); err != nil {
		return err
	}

	env, err := p.backend.CmdGetPodInfo(r.Form.Get("podName"))
	if err != nil {
		return err
	}

	return env.WriteJSON(w, http.StatusOK)
}

func (p *podRouter) getPodStats(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if err := httputils.ParseForm(r); err != nil {
		return err
	}

	env, err := p.backend.CmdGetPodStats(r.Form.Get("podId"))
	if err != nil {
		return err
	}

	return env.WriteJSON(w, http.StatusOK)
}

func (p *podRouter) getList(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if err := httputils.ParseForm(r); err != nil {
		return err
	}

	item := r.Form.Get("item")
	auxiliary := httputils.BoolValue(r, "auxiliary")
	pod := r.Form.Get("pod")
	vm := r.Form.Get("vm")

	glog.V(1).Infof("List type is %s, specified pod: [%s], specified vm: [%s], list auxiliary pod: %v", item, pod, vm, auxiliary)

	env, err := p.backend.CmdList(item, pod, vm, auxiliary)
	if err != nil {
		return err
	}

	return env.WriteJSON(w, http.StatusCreated)
}

func (p *podRouter) postPodCreate(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if err := httputils.ParseForm(r); err != nil {
		return err
	}

	if err := httputils.CheckForJSON(r); err != nil {
		return err
	}

	podArgs, _ := ioutil.ReadAll(r.Body)
	autoRemove := false
	if r.Form.Get("remove") == "yes" || r.Form.Get("remove") == "true" {
		autoRemove = true
	}
	glog.V(1).Infof("Args string is %s, autoremove %v", string(podArgs), autoRemove)

	env, err := p.backend.CmdCreatePod(string(podArgs), autoRemove)
	if err != nil {
		return err
	}

	return env.WriteJSON(w, http.StatusCreated)
}

func (p *podRouter) postPodLabels(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if err := httputils.ParseForm(r); err != nil {
		return err
	}

	podId := r.Form.Get("podId")
	labels := make(map[string]string)

	if err := json.Unmarshal([]byte(r.Form.Get("labels")), &labels); err != nil {
		return err
	}

	override := false
	if r.Form.Get("override") == "true" || r.Form.Get("override") == "yes" {
		override = true
	}

	env, err := p.backend.CmdSetPodLabels(podId, override, labels)
	if err != nil {
		return err
	}

	return env.WriteJSON(w, http.StatusCreated)
}

func (p *podRouter) postPodStart(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if err := httputils.ParseForm(r); err != nil {
		return err
	}

	podId := r.Form.Get("podId")
	vmId := r.Form.Get("vmId")
	tag := r.Form.Get("tag")

	if tag == "" {
		env, err := p.backend.CmdStartPod(nil, nil, podId, vmId, tag)
		if err != nil {
			return err
		}

		return env.WriteJSON(w, http.StatusOK)
	} else {
		// Setting up the streaming http interface.
		inStream, outStream, err := httputils.HijackConnection(w)
		if err != nil {
			return err
		}
		defer httputils.CloseStreams(inStream, outStream)

		fmt.Fprintf(outStream, "HTTP/1.1 101 UPGRADED\r\nContent-Type: application/vnd.docker.raw-stream\r\nConnection: Upgrade\r\nUpgrade: tcp\r\n\r\n")

		if _, err = p.backend.CmdStartPod(inStream, outStream.(io.WriteCloser), podId, vmId, tag); err != nil {
			return err
		}
		w.WriteHeader(http.StatusNoContent)
		return nil
	}
}

func (p *podRouter) postPodStop(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if err := httputils.ParseForm(r); err != nil {
		return err
	}

	podId := r.Form.Get("podId")
	stopVm := r.Form.Get("stopVm")

	env, err := p.backend.CmdStopPod(podId, stopVm)
	if err != nil {
		return err
	}

	return env.WriteJSON(w, http.StatusOK)
}

func (p *podRouter) postVmCreate(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if err := httputils.ParseForm(r); err != nil {
		return err
	}

	cpu, err := strconv.Atoi(r.Form.Get("cpu"))
	if err != nil {
		return err
	}
	mem, err := strconv.Atoi(r.Form.Get("mem"))
	if err != nil {
		return err
	}
	async := false
	if r.Form.Get("async") == "yes" || r.Form.Get("async") == "true" {
		async = true
	}

	env, err := p.backend.CmdCreateVm(cpu, mem, async)
	if err != nil {
		return err
	}

	return env.WriteJSON(w, http.StatusOK)
}

func (p *podRouter) deletePod(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if err := httputils.ParseForm(r); err != nil {
		return err
	}

	podId := r.Form.Get("podId")
	env, err := p.backend.CmdCleanPod(podId)
	if err != nil {
		return err
	}

	return env.WriteJSON(w, http.StatusOK)
}

func (p *podRouter) deleteVm(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if err := httputils.ParseForm(r); err != nil {
		return err
	}

	vmId := r.Form.Get("vm")
	env, err := p.backend.CmdKillVm(vmId)
	if err != nil {
		return err
	}

	return env.WriteJSON(w, http.StatusOK)
}
