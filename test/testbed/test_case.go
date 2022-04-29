package testbed

import (
	"log"
	"testing"
	"time"
)

// TestCase Running test case definition
type TestCase struct {
	t *testing.T

	// Agent process
	agent InfraRunner

	// Result validator
	validator ResourceValidator

	// errorSignal indicates an error in the test case execution, e.g. process execution
	// failure or exceeding resource consumption, etc. The actual error message is already
	// logged, this is only an indicator on which you can wait to be informed.
	errorSignal chan struct{}

	// errorCause keeps the latest generated error
	errorCause error
}

func NewTestCase(t *testing.T, a InfraRunner, v ResourceValidator) *TestCase {
	return &TestCase{
		t:           t,
		agent:       a,
		validator:   v,
		errorSignal: make(chan struct{}),
	}
}

func (tc *TestCase) StartInfraAgent() {

	err := tc.agent.Start()
	if err != nil {
		tc.handleError(err)
		return
	}

	// Start watching resource consumption
	go func() {
		err := tc.agent.WatchResourceConsumption()
		if err != nil {
			tc.handleError(err)
			return
		}
	}()

}

func (tc *TestCase) StopInfraAgent() {
	rc := tc.agent.GetResourceConsumption()
	result := tc.validator.Report(rc)
	log.Println(result)

	err := tc.agent.Stop()
	if err != nil {
		tc.handleError(err)
		return
	}

	err = tc.validator.Validate(rc)
	if err != nil {
		tc.handleError(err)
	}
}

func (tc *TestCase) Sleep(d time.Duration) {
	select {
	case <-time.After(d):
	case <-tc.errorSignal:
	}
}

func (tc *TestCase) handleError(err error) {
	log.Println(err)

	tc.t.Error(err)

	tc.errorCause = err

	close(tc.errorSignal)
}
