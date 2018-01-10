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
	APIServer ControlPlaneProcess
	Etcd      ControlPlaneProcess

	apiServerSession SimpleSession
	etcdSession      SimpleSession
}

type ControlPlaneProcess interface {
	URL() (*url.URL, error)
	Command() (*exec.Cmd, error)
	CleanUp() error
	UpMessage() string
}

//go:generate counterfeiter . ControlPlaneProcess

type SimpleSession interface {
	Terminate() *gexec.Session
}

// NewControlPlane will give you a ControlPlane struct that's properly wired together.
func NewControlPlane() *ControlPlane {
	return &ControlPlane{
		Etcd:      &Etcd{},
		APIServer: &APIServer{},
	}
}

func (f *ControlPlane) Start() error {
	var err error

	f.etcdSession, err = startProcess(f.Etcd)
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

	f.apiServerSession, err = startProcess(f.APIServer)
	if err != nil {
		return err
	}

	return nil
}

func (f *ControlPlane) Stop() error {
	if err := stopProcess(f.apiServerSession); err != nil {
		return err
	}
	if err := f.APIServer.CleanUp(); err != nil {
		return err
	}

	if err := stopProcess(f.etcdSession); err != nil {
		return err
	}
	if err := f.Etcd.CleanUp(); err != nil {
		return err
	}

	return nil
}

// APIServerURL returns the URL to the APIServer. Clients can use this URL to connect to the APIServer.
func (f *ControlPlane) APIServerURL() (*url.URL, error) {
	return f.APIServer.URL()
}

func startProcess(p ControlPlaneProcess) (SimpleSession, error) {
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

	select {
	case <-detectedStart:
		return session, nil
	case <-timedOut:
		return nil, fmt.Errorf("timeout waiting for XXX to start serving")
	}
}

func stopProcess(session SimpleSession) error {
	if session == nil {
		return nil
	}

	detectedStop := session.Terminate().Exited
	timedOut := time.After(20 * time.Second)

	select {
	case <-detectedStop:
		return nil
	case <-timedOut:
		return fmt.Errorf("timeout waiting for XXX to stop")
	}
}
