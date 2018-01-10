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

//go:generate counterfeiter . ControlPlaneProcess
type ControlPlaneProcess interface {
	URL() (*url.URL, error)
	Command() (*exec.Cmd, error)
	CleanUp() error
	UpMessage() string
}

//go:generate counterfeiter . ProcessStarter
type ProcessStarter func(process ControlPlaneProcess) (ProcessStopper, error)

//go:generate counterfeiter . ProcessStopper
type ProcessStopper func() error

func StartProcess(p ControlPlaneProcess) (ProcessStopper, error) {
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
