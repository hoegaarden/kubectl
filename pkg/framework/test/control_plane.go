// Package test an integration test framework for k8s
package test

import (
	"net/url"
)

// ControlPlane is a struct that knows how to start your test control plane.
//
// Right now, that means Etcd and your APIServer. This is likely to increase in future.
type ControlPlane struct {
	APIServer ControlPlaneProcess
	Etcd      ControlPlaneProcess
}

// ControlPlaneProcess knows how to start and stop a ControlPlane process.
// This interface is potentially going to be expanded to e.g. allow access to the processes StdOut/StdErr
// and other internals.
type ControlPlaneProcess interface {
	Start() error
	Stop() error
	URL() (string, error)
}

//go:generate counterfeiter . ControlPlaneProcess

// NewControlPlane will give you a ControlPlane struct that's properly wired together.
func NewControlPlane() *ControlPlane {
	etcd := &Etcd{}

	etcdURL, err := etcd.URL()
	if err != nil {
		panic(err)
	}
	parsedEtcdURL, err := url.Parse(etcdURL)
	if err != nil {
		panic(err)
	}

	apiServer := &APIServer{
		EtcdAddress: parsedEtcdURL,
	}
	return &ControlPlane{
		APIServer: apiServer,
		Etcd:      etcd,
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
func (f *ControlPlane) APIServerURL() (string, error) {
	return f.APIServer.URL()
}
