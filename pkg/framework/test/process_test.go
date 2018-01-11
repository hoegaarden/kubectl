package test_test

import (
	"fmt"
	"os/exec"

	. "k8s.io/kubectl/pkg/framework/test"
	"k8s.io/kubectl/pkg/framework/test/testfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("StartProcess", func() {
	var (
		fakeProcess *testfakes.FakeControlPlaneProcess
	)
	BeforeEach(func() {
		fakeProcess = &testfakes.FakeControlPlaneProcess{}
	})

	It("can start & stop a process", func() {
		fakeProcess.CommandReturns(getTestCommand(), nil)
		fakeProcess.UpMessageReturns("5")

		stopper, err := StartProcess(fakeProcess)
		Expect(err).NotTo(HaveOccurred())

		Expect(stopper()).To(Succeed())
	})

	Describe("when the process cannot create a command", func() {
		It("propagates the error", func() {
			fakeProcess.CommandReturns(nil, fmt.Errorf("some error"))

			stopper, err := StartProcess(fakeProcess)
			Expect(err).To(MatchError("some error"))
			Expect(stopper).To(BeNil())
		})
	})

	Describe("when the command fails to start", func() {
		It("propagates the error", func() {
			fakeProcess.CommandReturns(&exec.Cmd{Path: "/nonexistent"}, nil)

			stopper, err := StartProcess(fakeProcess)
			Expect(err).To(MatchError("fork/exec /nonexistent: no such file or directory"))
			Expect(stopper).To(BeNil())
		})
	})

	// When the process terminates right away
	// When the process runs into a timeout while starting
	// When the process runs into a timeout while stopping
	// When CleanUp is nil / is not nil
	// When UpMessage is "" / not ""
})

func getTestCommand() *exec.Cmd {
	return exec.Command(
		"bash", "-c",
		`
			i=1
			while true
			do
				echo $i >&2 ; let 'i += 1' ; sleep 0.1
			done
		`,
	)
}
