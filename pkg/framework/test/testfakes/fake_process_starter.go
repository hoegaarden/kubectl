// Code generated by counterfeiter. DO NOT EDIT.
package testfakes

import (
	"sync"

	"k8s.io/kubectl/pkg/framework/test"
)

type FakeProcessStarter struct {
	Stub        func(process test.ControlPlaneProcess) (test.ProcessStopper, error)
	mutex       sync.RWMutex
	argsForCall []struct {
		process test.ControlPlaneProcess
	}
	returns struct {
		result1 test.ProcessStopper
		result2 error
	}
	returnsOnCall map[int]struct {
		result1 test.ProcessStopper
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeProcessStarter) Spy(process test.ControlPlaneProcess) (test.ProcessStopper, error) {
	fake.mutex.Lock()
	ret, specificReturn := fake.returnsOnCall[len(fake.argsForCall)]
	fake.argsForCall = append(fake.argsForCall, struct {
		process test.ControlPlaneProcess
	}{process})
	fake.recordInvocation("ProcessStarter", []interface{}{process})
	fake.mutex.Unlock()
	if fake.Stub != nil {
		return fake.Stub(process)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fake.returns.result1, fake.returns.result2
}

func (fake *FakeProcessStarter) CallCount() int {
	fake.mutex.RLock()
	defer fake.mutex.RUnlock()
	return len(fake.argsForCall)
}

func (fake *FakeProcessStarter) ArgsForCall(i int) test.ControlPlaneProcess {
	fake.mutex.RLock()
	defer fake.mutex.RUnlock()
	return fake.argsForCall[i].process
}

func (fake *FakeProcessStarter) Returns(result1 test.ProcessStopper, result2 error) {
	fake.Stub = nil
	fake.returns = struct {
		result1 test.ProcessStopper
		result2 error
	}{result1, result2}
}

func (fake *FakeProcessStarter) ReturnsOnCall(i int, result1 test.ProcessStopper, result2 error) {
	fake.Stub = nil
	if fake.returnsOnCall == nil {
		fake.returnsOnCall = make(map[int]struct {
			result1 test.ProcessStopper
			result2 error
		})
	}
	fake.returnsOnCall[i] = struct {
		result1 test.ProcessStopper
		result2 error
	}{result1, result2}
}

func (fake *FakeProcessStarter) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.mutex.RLock()
	defer fake.mutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeProcessStarter) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ test.ProcessStarter = new(FakeProcessStarter).Spy
