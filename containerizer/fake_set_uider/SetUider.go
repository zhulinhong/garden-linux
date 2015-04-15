// This file was generated by counterfeiter
package fake_set_uider

import (
	"sync"

	"github.com/cloudfoundry-incubator/garden-linux/containerizer"
)

type FakeSetUider struct {
	SetUidStub        func() error
	setUidMutex       sync.RWMutex
	setUidArgsForCall []struct{}
	setUidReturns struct {
		result1 error
	}
}

func (fake *FakeSetUider) SetUid() error {
	fake.setUidMutex.Lock()
	fake.setUidArgsForCall = append(fake.setUidArgsForCall, struct{}{})
	fake.setUidMutex.Unlock()
	if fake.SetUidStub != nil {
		return fake.SetUidStub()
	} else {
		return fake.setUidReturns.result1
	}
}

func (fake *FakeSetUider) SetUidCallCount() int {
	fake.setUidMutex.RLock()
	defer fake.setUidMutex.RUnlock()
	return len(fake.setUidArgsForCall)
}

func (fake *FakeSetUider) SetUidReturns(result1 error) {
	fake.SetUidStub = nil
	fake.setUidReturns = struct {
		result1 error
	}{result1}
}

var _ containerizer.SetUider = new(FakeSetUider)
