package test_test

import (
	"io"
	"os/exec"

	. "k8s.io/kubectl/pkg/framework/test"

	"fmt"

	"time"

	"net/url"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"k8s.io/kubectl/pkg/framework/test/testfakes"
)

var _ = Describe("Apiserver", func() {
	var (
		fakeSession      *testfakes.FakeSimpleSession
		apiServer        *APIServer
		apiServerStopper chan struct{}
		cleanupCallCount int
	)

	BeforeEach(func() {
		fakeSession = &testfakes.FakeSimpleSession{}

		apiServerStopper = make(chan struct{}, 1)
		fakeSession.TerminateReturns(&gexec.Session{
			Exited: apiServerStopper,
		})
		close(apiServerStopper)

		apiServer = &APIServer{
			Address: &url.URL{Scheme: "http", Host: "the.host.for.api.server:5678"},
			Path:    "/some/path/to/apiserver",
			CertDir: &CleanableDirectory{
				Path: "/some/path/to/certdir",
				Cleanup: func() error {
					cleanupCallCount += 1
					return nil
				},
			},
			EtcdAddress: &url.URL{Host: "the.etcd.url"},
			StopTimeout: 500 * time.Millisecond,
		}
	})

	Describe("starting and stopping the server", func() {
		Context("when given a path to a binary that runs for a long time", func() {
			It("can start and stop that binary", func() {
				fakeSession.ExitCodeReturnsOnCall(0, -1)
				fakeSession.ExitCodeReturnsOnCall(1, 143)

				apiServer.Address = &url.URL{Scheme: "http", Host: "this.is.the.API.server:1234"}

				apiServer.ProcessStarter = func(command *exec.Cmd, out, err io.Writer) (SimpleSession, error) {
					Expect(command.Args).To(ContainElement("--insecure-port=1234"))
					Expect(command.Args).To(ContainElement("--insecure-bind-address=this.is.the.API.server"))
					Expect(command.Args).To(ContainElement("--etcd-servers=//the.etcd.url"))
					Expect(command.Args).To(ContainElement("--cert-dir=/some/path/to/certdir"))
					Expect(command.Path).To(Equal("/some/path/to/apiserver"))
					fmt.Fprint(err, "Serving insecurely on this.is.the.API.server:1234")
					return fakeSession, nil
				}

				By("Starting the API Server")
				err := apiServer.Start()
				Expect(err).NotTo(HaveOccurred())

				By("Stopping the API Server")
				Expect(apiServer.Stop()).To(Succeed())

				Expect(cleanupCallCount).To(Equal(1))
				Expect(fakeSession.TerminateCallCount()).To(Equal(1))
			})
		})

		Context("when the certificate directory cannot be destroyed", func() {
			It("propagates the error", func() {
				apiServer.CertDir.Cleanup = func() error { return fmt.Errorf("destroy failed") }
				apiServer.Address = &url.URL{Scheme: "http", Host: "this.is.apiserver:1234"}
				apiServer.ProcessStarter = func(Command *exec.Cmd, out, err io.Writer) (SimpleSession, error) {
					fmt.Fprint(err, "Serving insecurely on this.is.apiserver:1234")
					return fakeSession, nil
				}

				Expect(apiServer.Start()).To(Succeed())
				err := apiServer.Stop()
				Expect(err).To(MatchError(ContainSubstring("destroy failed")))
			})
		})

		Context("when there is no function to cleanup the certificate directory", func() {
			It("does not panic", func() {
				apiServer.CertDir.Cleanup = nil
				apiServer.Address = &url.URL{Scheme: "http", Host: "this.is.apiserver:1234"}
				apiServer.ProcessStarter = func(Command *exec.Cmd, out, err io.Writer) (SimpleSession, error) {
					fmt.Fprint(err, "Serving insecurely on this.is.apiserver:1234")
					return fakeSession, nil
				}

				Expect(apiServer.Start()).To(Succeed())

				var err error
				Expect(func() {
					err = apiServer.Stop()
				}).NotTo(Panic())
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when EtcdAddress is not specified", func() {
			It("propagates the error", func() {
				apiServer.EtcdAddress = nil
				apiServer.ProcessStarter = func(Command *exec.Cmd, out, err io.Writer) (SimpleSession, error) {
					Expect(true).To(BeFalse(),
						"the api server process starter shouldn't be called if etcd's URL is nil")
					return nil, nil
				}

				err := apiServer.Start()
				Expect(err).To(MatchError(ContainSubstring("Etcd URL cannot be empty")))
			})
		})

		Context("when the starter returns an error", func() {
			It("propagates the error", func() {
				apiServer.ProcessStarter = func(command *exec.Cmd, out, err io.Writer) (SimpleSession, error) {
					return nil, fmt.Errorf("Some error in the apiserver starter.")
				}

				err := apiServer.Start()
				Expect(err).To(MatchError(ContainSubstring("Some error in the apiserver starter.")))
			})
		})

		Context("when the starter takes longer than our timeout", func() {
			It("gives us a timeout error", func() {
				apiServer.StartTimeout = 1 * time.Nanosecond
				apiServer.ProcessStarter = func(command *exec.Cmd, out, err io.Writer) (SimpleSession, error) {
					return &gexec.Session{}, nil
				}

				err := apiServer.Start()
				Expect(err).To(MatchError(ContainSubstring("timeout waiting for apiserver to start serving")))
			})
		})

		Context("when we try to stop a server that hasn't been started", func() {
			It("is a noop and does not call exit on the session", func() {
				apiServer.ProcessStarter = func(command *exec.Cmd, out, err io.Writer) (SimpleSession, error) {
					return fakeSession, nil
				}
				apiServer.Stop()
				Expect(fakeSession.ExitCodeCallCount()).To(Equal(0))
			})
		})

		Context("when Stop() times out", func() {
			JustBeforeEach(func() {
				apiServerStopperWillNeverBeUsed := make(chan struct{}, 1)
				fakeSession.TerminateReturns(&gexec.Session{
					Exited: apiServerStopperWillNeverBeUsed,
				})
			})
			It("propagates the error", func() {
				apiServer.Address = &url.URL{Scheme: "http", Host: "this.is.apiserver:1234"}
				apiServer.ProcessStarter = func(Command *exec.Cmd, out, err io.Writer) (SimpleSession, error) {
					fmt.Fprint(err, "Serving insecurely on this.is.apiserver:1234")
					return fakeSession, nil
				}

				Expect(apiServer.Start()).To(Succeed())
				err := apiServer.Stop()
				Expect(err).To(MatchError(ContainSubstring("timeout")))
			})
		})
	})

	Describe("querying the server for its URL", func() {
		It("can be queried for the URL it listens on", func() {
			apiServerURL, err := apiServer.URL()
			Expect(err).NotTo(HaveOccurred())
			Expect(apiServerURL).To(Equal("http://the.host.for.api.server:5678"))
		})
	})
})
