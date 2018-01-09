package test

import (
	"os/exec"

	"fmt"
	"net/url"

	"k8s.io/kubectl/pkg/framework/test/internal"
)

// Etcd knows how to run an etcd server.
//
// The documentation and examples for the Etcd's properties can be found in
// in the documentation for the `APIServer`, as both implement a `ControlPaneProcess`.
type Etcd struct {
	Address *url.URL
	Path    string
	DataDir *CleanableDirectory
}

// URL returns the URL Etcd is listening on. Clients can use this to connect to Etcd.
func (e *Etcd) URL() (*url.URL, error) {
	if err := e.ensureInitialized(); err != nil {
		return nil, err
	}
	return e.Address, nil
}

// Command bla
func (e *Etcd) Command() (*exec.Cmd, error) {
	err := e.ensureInitialized()
	if err != nil {
		return nil, err
	}

	args := []string{
		"--debug",
		"--listen-peer-urls=http://localhost:0",
		fmt.Sprintf("--advertise-client-urls=%s", e.Address),
		fmt.Sprintf("--listen-client-urls=%s", e.Address),
		fmt.Sprintf("--data-dir=%s", e.DataDir.Path),
	}

	cmd := exec.Command(e.Path, args...)

	return cmd, nil
}

func (e *Etcd) UpMessage() string {
	return "erving insecure"
}

// CleanUp blupp
func (e *Etcd) CleanUp() error {
	if e.DataDir.Cleanup == nil {
		return nil
	}
	return e.DataDir.Cleanup()
}

func (e *Etcd) ensureInitialized() error {
	if e.Path == "" {
		e.Path = internal.BinPathFinder("etcd")
	}
	if e.Address == nil {
		am := &internal.AddressManager{}
		port, host, err := am.Initialize()
		if err != nil {
			return err
		}
		e.Address = &url.URL{
			Scheme: "http",
			Host:   fmt.Sprintf("%s:%d", host, port),
		}
	}
	if e.DataDir == nil {
		dataDir, err := newDirectory()
		if err != nil {
			return err
		}
		e.DataDir = dataDir
	}

	return nil
}
