// Package test an integration test framework for k8s
package test

import (
	"net/url"

	"fmt"

	"k8s.io/kubectl/pkg/framework/test/internal"
)

// ControlPlane is a struct that knows how to start your test control plane.
//
// Right now, that means Etcd and your APIServer. This is likely to increase in future.
type ControlPlane struct {
	Starter ProcessStarter

	APIServer ControlPlaneProcess
	Etcd      ControlPlaneProcess

	apiServerStopper ProcessStopper
	etcdStopper      ProcessStopper
}

func NewControlPlane() (*ControlPlane, error) {
	am := internal.AddressManager{}
	etcdPort, etcdHost, err := am.Initialize()
	if err != nil {
		return nil, err
	}
	etcdURL := &url.URL{
		Host:   fmt.Sprintf("%s:%d", etcdHost, etcdPort),
		Scheme: "http",
	}

	return &ControlPlane{
		Etcd: &Etcd{
			Address: etcdURL,
		},
		APIServer: &APIServer{
			EtcdAddress: etcdURL,
		},
	}, nil
}

func (f *ControlPlane) Start() error {
	f.ensureInitialized()

	var err error

	f.etcdStopper, err = f.Starter(f.Etcd)
	if err != nil {
		return err
	}

	f.apiServerStopper, err = f.Starter(f.APIServer)
	if err != nil {
		return err
	}

	return nil
}

func (f *ControlPlane) Stop() error {
	if f.apiServerStopper != nil {
		if err := f.apiServerStopper(); err != nil {
			return err
		}
	}
	if f.etcdStopper != nil {
		if err := f.etcdStopper(); err != nil {
			return err
		}
	}
	return nil
}

// APIServerURL returns the URL to the APIServer. Clients can use this URL to connect to the APIServer.
func (f *ControlPlane) APIServerURL() (*url.URL, error) {
	return f.APIServer.URL()
}

func (f *ControlPlane) ensureInitialized() {
	if f.Starter == nil {
		f.Starter = StartProcess
	}
	if f.Etcd == nil {
		f.Etcd = &Etcd{}
	}
	if f.APIServer == nil {
		f.APIServer = &APIServer{}
	}
}
