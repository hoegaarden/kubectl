package test_test

import (
	. "k8s.io/kubectl/pkg/framework/test"

	"fmt"

	"net/url"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/kubectl/pkg/framework/test/testfakes"
)

var _ = Describe("ControlPlane", func() {
	Context("with a properly configured set of ControlPlaneProcesses", func() {
		var (
			fakeAPIServerProcess *testfakes.FakeControlPlaneProcess
			fakeEtcdProcess      *testfakes.FakeControlPlaneProcess
			fakeStarter          *testfakes.FakeProcessStarter
			controlPlane         ControlPlane
		)
		BeforeEach(func() {
			fakeAPIServerProcess = &testfakes.FakeControlPlaneProcess{}
			fakeEtcdProcess = &testfakes.FakeControlPlaneProcess{}
			fakeStarter = &testfakes.FakeProcessStarter{}

			succesfulStopper := createStopper("")
			fakeStarter.Returns(succesfulStopper, nil)

			controlPlane = ControlPlane{
				Starter:   fakeStarter.Spy,
				APIServer: fakeAPIServerProcess,
				Etcd:      fakeEtcdProcess,
			}
		})

		It("can start the ControlPlane", func() {
			Expect(controlPlane.Start()).To(Succeed())
			Expect(fakeStarter.CallCount()).To(Equal(2))
		})

		It("does not panic when stopping the ControlPlane without starting it", func() {
			Expect(controlPlane.Stop()).To(Succeed())
			Expect(fakeStarter.CallCount()).To(Equal(0))
		})

		Context("when starting a ControlPlaneProcess fails", func() {
			It("propagates the error", func() {
				fakeStarter.Returns(nil, fmt.Errorf("another error"))
				err := controlPlane.Start()
				Expect(err).To(MatchError(ContainSubstring("another error")))
			})
		})

		Context("when stopping a ControlPlaneProcess fails", func() {
			It("propagates the error", func() {
				failingStopper := createStopper("error on stop")
				fakeStarter.Returns(failingStopper, nil)

				Expect(controlPlane.Start()).To(Succeed())
				err := controlPlane.Stop()
				Expect(err).To(MatchError(ContainSubstring("error on stop")))
			})
		})

		It("can be queried for the APIServer URL", func() {
			fakeAPIServerProcess.URLReturns(&url.URL{Host: "some.url.to.the.apiserver"}, nil)

			url, err := controlPlane.APIServerURL()
			Expect(err).NotTo(HaveOccurred())
			Expect(url.String()).To(Equal("//some.url.to.the.apiserver"))
		})

		Context("when querying the URL fails", func() {
			It("propagates the error", func() {
				fakeAPIServerProcess.URLReturns(nil, fmt.Errorf("URL error"))
				_, err := controlPlane.APIServerURL()
				Expect(err).To(MatchError(ContainSubstring("URL error")))
			})
		})
	})
})

func createStopper(errorMessage string) func() error {
	var err error
	if errorMessage != "" {
		err = fmt.Errorf(errorMessage)
	}
	return func() error {
		return err
	}
}
