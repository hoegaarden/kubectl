package test

import (
	"fmt"
	"os/exec"

	"net/url"

	"k8s.io/kubectl/pkg/framework/test/internal"
)

// APIServer knows how to run a kubernetes apiserver.
type APIServer struct {
	// Address is the address, a host and a port, the ApiServer should listen on for client connections.
	// If this is not specified, we default to a random free port on localhost.
	Address *url.URL

	// Path is the path to the apiserver binary. If this is left as the empty
	// string, we will attempt to locate a binary, by checking for the
	// TEST_ASSET_KUBE_APISERVER environment variable, and the default test
	// assets directory.
	Path string

	// CertDir is a struct holding a path to a certificate directory and a function to cleanup that directory.
	CertDir *CleanableDirectory

	// EtcdAddress points to an Etcd we can use to store APIServer's state
	EtcdAddress *url.URL
}

// URL returns the URL APIServer is listening on. Clients can use this to connect to APIServer.
func (s *APIServer) URL() (*url.URL, error) {
	if err := s.ensureInitialized(); err != nil {
		return nil, err
	}
	return s.Address, nil
}

// CleanUp bla
func (s *APIServer) CleanUp() error {
	if s.CertDir.Cleanup == nil {
		return nil
	}
	return s.CertDir.Cleanup()
}

// Command blupp
func (s *APIServer) Command() (*exec.Cmd, error) {
	err := s.ensureInitialized()
	if err != nil {
		return nil, err
	}

	args := []string{
		"--authorization-mode=Node,RBAC",
		"--runtime-config=admissionregistration.k8s.io/v1alpha1",
		"--v=3", "--vmodule=",
		"--admission-control=Initializers,NamespaceLifecycle,LimitRanger,ServiceAccount,SecurityContextDeny,DefaultStorageClass,DefaultTolerationSeconds,GenericAdmissionWebhook,ResourceQuota",
		"--admission-control-config-file=",
		"--bind-address=0.0.0.0",
		"--storage-backend=etcd3",
		fmt.Sprintf("--etcd-servers=%s", s.EtcdAddress.String()),
		fmt.Sprintf("--cert-dir=%s", s.CertDir.Path),
		fmt.Sprintf("--insecure-port=%s", s.Address.Port()),
		fmt.Sprintf("--insecure-bind-address=%s", s.Address.Hostname()),
	}

	cmd := exec.Command(s.Path, args...)

	return cmd, nil
}

func (s *APIServer) UpMessage() string {
	return "erving insecure"
}

func (s *APIServer) ensureInitialized() error {
	if s.Path == "" {
		s.Path = internal.BinPathFinder("kube-apiserver")
	}
	if s.Address == nil {
		am := &internal.AddressManager{}
		port, host, err := am.Initialize()
		if err != nil {
			return err
		}
		s.Address = &url.URL{
			Scheme: "http",
			Host:   fmt.Sprintf("%s:%d", host, port),
		}
	}
	if s.CertDir == nil {
		certDir, err := newDirectory()
		if err != nil {
			return err
		}
		s.CertDir = certDir
	}
	if s.EtcdAddress == nil {
		return fmt.Errorf("Etcd URL cannot be empty")
	}

	return nil
}
