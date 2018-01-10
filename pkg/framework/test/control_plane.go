// Package test an integration test framework for k8s
package test

import (
	"fmt"
	"net/url"
	"os/exec"
	"path"
	"time"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

// ControlPlane is a struct that knows how to start your test control plane.
//
// Right now, that means Etcd and your APIServer. This is likely to increase in future.
type ControlPlane struct {
	APIServer ControlPlaneProcess
	Etcd      ControlPlaneProcess

	apiServerStopper Stopper
	etcdStopper      Stopper
}

type Stopper func() error

type ControlPlaneProcess interface {
	URL() (*url.URL, error)
	Command() (*exec.Cmd, error)
	CleanUp() error
	UpMessage() string
}

//go:generate counterfeiter . ControlPlaneProcess

func (f *ControlPlane) Start() error {
	f.ensureInitialized()

	var err error

	f.etcdStopper, err = startProcess(f.Etcd)
	if err != nil {
		return err
	}

	etcdUrl, err := f.Etcd.URL()
	if err != nil {
		return err
	}

	// TODO ~!@#$%^
	if f.APIServer.(*APIServer).EtcdAddress == nil {
		f.APIServer.(*APIServer).EtcdAddress = etcdUrl
	}

	f.apiServerStopper, err = startProcess(f.APIServer)
	if err != nil {
		return err
	}

	return nil
}

func (f *ControlPlane) Stop() error {
	if err := f.apiServerStopper(); err != nil {
		return err
	}
	if err := f.etcdStopper(); err != nil {
		return err
	}
	return nil
}

// APIServerURL returns the URL to the APIServer. Clients can use this URL to connect to the APIServer.
func (f *ControlPlane) APIServerURL() (*url.URL, error) {
	return f.APIServer.URL()
}

func (f *ControlPlane) ensureInitialized() {
	if f.Etcd == nil {
		f.Etcd = &Etcd{}
	}
	if f.APIServer == nil {
		f.APIServer = &APIServer{}
	}
}

func startProcess(p ControlPlaneProcess) (Stopper, error) {
	command, err := p.Command()
	if err != nil {
		return nil, err
	}

	stdErr := gbytes.NewBuffer()
	detectedStart := stdErr.Detect(p.UpMessage())
	timedOut := time.After(20 * time.Second)

	session, err := gexec.Start(command, nil, stdErr)
	if err != nil {
		return nil, err
	}

	binName := getBinName(command)
	stopper := func() error {
		if session == nil {
			return nil
		}

		detectedStop := session.Terminate().Exited
		timedOut := time.After(20 * time.Second)

		select {
		case <-detectedStop:
			return p.CleanUp()
		case <-timedOut:
			return fmt.Errorf("timeout waiting for %s to stop", binName)
		}
	}

	select {
	case <-detectedStart:
		return stopper, nil
	case <-timedOut:
		return nil, fmt.Errorf("timeout waiting for %s to start serving", binName)
	}
}

// getBinName is just a helper to extract a nice name from a *exec.Command
func getBinName(cmd *exec.Cmd) string {
	name := path.Base(cmd.Path)
	if name == "." || name == "/" {
		name = "<unknown>"
	}
	return name
}
