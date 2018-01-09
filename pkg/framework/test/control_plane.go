// Package test an integration test framework for k8s
package test

import (
	"fmt"
	"net/url"
	"os/exec"
	"time"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

// ControlPlane is a struct that knows how to start your test control plane.
//
// Right now, that means Etcd and your APIServer. This is likely to increase in future.
type ControlPlane struct {
	APIServer ProcessManager
	Etcd      ProcessManager
}

type ProcessManager interface {
	Start() error
	Stop() error
	URL() (*url.URL, error)
}

type ControlPlaneProcess interface {
	URL() (*url.URL, error)
	Command() (*exec.Cmd, error)
	CleanUp() error
	UpMessage() string
}

//go:generate counterfeiter . ControlPlaneProcess

type process struct {
	Implementation ControlPlaneProcess

	session      *gexec.Session // SimpleSession
	stdErr       *gbytes.Buffer
	stdOut       *gbytes.Buffer
	startTimeout time.Duration
	stopTimeout  time.Duration
}

// NewControlPlane will give you a ControlPlane struct that's properly wired together.
func NewControlPlane() *ControlPlane {
	etcdProc := &process{
		Implementation: &Etcd{},
	}

	etcdURL, err := etcdProc.Implementation.URL()
	if err != nil {
		panic(err) // TODO me no like panic ...
	}

	apiServerProc := &process{
		Implementation: &APIServer{
			EtcdAddress: etcdURL,
		},
	}

	return &ControlPlane{
		APIServer: apiServerProc,
		Etcd:      etcdProc,
	}
}

// Start will start your control plane. To stop it, call Stop().
func (f *ControlPlane) Start() error {
	if err := f.Etcd.Start(); err != nil {
		return err
	}
	return f.APIServer.Start()
}

// Stop will stop your control plane, and clean up their data.
func (f *ControlPlane) Stop() error {
	if err := f.APIServer.Stop(); err != nil {
		return err
	}
	return f.Etcd.Stop()
}

// APIServerURL returns the URL to the APIServer. Clients can use this URL to connect to the APIServer.
func (f *ControlPlane) APIServerURL() (*url.URL, error) {
	return f.APIServer.URL()
}

func (pm *process) ensureInitialized() {
	if pm.stdErr == nil {
		pm.stdErr = gbytes.NewBuffer()
	}
	if pm.stdOut == nil {
		pm.stdOut = gbytes.NewBuffer()
	}

	if pm.startTimeout == 0 {
		pm.startTimeout = 20 * time.Second
	}
	if pm.stopTimeout == 0 {
		pm.stopTimeout = 20 * time.Second
	}
}

func (pm *process) Start() error {
	pm.ensureInitialized()

	command, err := pm.Implementation.Command()
	if err != nil {
		return err
	}

	detectedStart := pm.stdErr.Detect(fmt.Sprintf(pm.Implementation.UpMessage()))
	timedOut := time.After(pm.startTimeout)

	pm.session, err = gexec.Start(command, pm.stdOut, pm.stdErr)
	if err != nil {
		return err
	}

	select {
	case <-detectedStart:
		return nil
	case <-timedOut:
		return fmt.Errorf("timeout waiting for etcd to start serving")
	}
}

func (pm *process) Stop() error {
	if pm.session == nil {
		return nil
	}

	session := pm.session.Terminate()
	detectedStop := session.Exited
	timedOut := time.After(pm.stopTimeout)

	select {
	case <-detectedStop:
		break
	case <-timedOut:
		return fmt.Errorf("timeout waiting for XXX to stop")
	}

	if pm.Implementation.CleanUp == nil {
		return nil
	}
	return pm.Implementation.CleanUp()
}

func (pm *process) URL() (*url.URL, error) {
	return pm.Implementation.URL()
}
