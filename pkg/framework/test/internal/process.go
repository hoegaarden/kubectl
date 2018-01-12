package internal

import (
	"fmt"
	"net/url"
	"os/exec"
	"time"

	"os"

	"io/ioutil"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

type CommonStuff struct {
	URL              *url.URL
	Dir              string
	DirNeedsCleaning bool
	Path             string
	StopTimeout      time.Duration
	StartTimeout     time.Duration
}

func Start(command *exec.Cmd, startMessage string, startTimeout time.Duration) (*gexec.Session, error) {
	stdErr := gbytes.NewBuffer()
	detectedStart := stdErr.Detect(startMessage)
	timedOut := time.After(startTimeout)

	session, err := gexec.Start(command, gbytes.NewBuffer(), stdErr)
	if err != nil {
		return session, err
	}

	select {
	case <-detectedStart:
		return session, nil
	case <-timedOut:
		return session, fmt.Errorf("timeout waiting for apiserver to start serving")
	}

}

func Stop(session *gexec.Session, stopTimeout time.Duration, dirToClean string, dirNeedsCleaning bool) error {
	if session == nil {
		return nil
	}

	detectedStop := session.Terminate().Exited
	timedOut := time.After(stopTimeout)

	select {
	case <-detectedStop:
		break
	case <-timedOut:
		return fmt.Errorf("timeout waiting for etcd to stop")
	}

	if dirNeedsCleaning {
		return os.RemoveAll(dirToClean)
	}

	return nil
}

func NewCommonStuff(
	symbolicName string,
	path string,
	listenURL *url.URL,
	dir string,
	startTimeout time.Duration,
	stopTimeout time.Duration,
) (CommonStuff, error) {
	common := CommonStuff{
		Path:             path,
		URL:              listenURL,
		Dir:              dir,
		DirNeedsCleaning: false,
		StartTimeout:     startTimeout,
		StopTimeout:      stopTimeout,
	}

	if path == "" {
		common.Path = BinPathFinder(symbolicName)
	}

	if listenURL == nil {
		am := &AddressManager{}
		port, host, err := am.Initialize()
		if err != nil {
			return CommonStuff{}, err
		}
		common.URL = &url.URL{
			Scheme: "http",
			Host:   fmt.Sprintf("%s:%d", host, port),
		}
	}

	if dir == "" {
		newDir, err := ioutil.TempDir("", "k8s_test_framework_")
		if err != nil {
			return CommonStuff{}, err
		}
		common.Dir = newDir
		common.DirNeedsCleaning = true
	}

	if stopTimeout == 0 {
		common.StopTimeout = 20 * time.Second
	}

	if startTimeout == 0 {
		common.StartTimeout = 20 * time.Second
	}

	return common, nil
}