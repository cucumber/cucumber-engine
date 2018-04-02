package runner

import (
	"github.com/cucumber/cucumber-pickle-runner/src/dto"
	"github.com/cucumber/cucumber-pickle-runner/src/dto/event"
	gherkin "github.com/cucumber/gherkin-go"
)

// NewTestCaseRunnerOptions are the options for NewTestCaseRunner
type NewTestCaseRunnerOptions struct {
	Pickle                        *gherkin.Pickle
	URI                           string
	SendCommand                   func(*dto.Command)
	SendCommandAndAwaitResponse   func(*dto.Command) *dto.Command
	BeforeTestCaseHookDefinitions []*dto.TestCaseHookDefinition
	AfterTestCaseHookDefinitions  []*dto.TestCaseHookDefinition
	StepDefinitions               []*dto.StepDefinition
}

// TestCaseRunner runs a test case
type TestCaseRunner struct {
	afterTestCaseHookDefinitions  []*dto.TestCaseHookDefinition
	beforeTestCaseHookDefinitions []*dto.TestCaseHookDefinition
	pickle                        *gherkin.Pickle
	sendCommand                   func(*dto.Command)
	sendCommandAndAwaitResponse   func(*dto.Command) *dto.Command
	stepIndexToStepDefinitions    [][]*dto.StepDefinition
	uri                           string

	result *dto.TestResult
}

// NewTestCaseRunner returns a TestCaseRunner
func NewTestCaseRunner(opts *NewTestCaseRunnerOptions) *TestCaseRunner {
	stepIndexToStepDefinitions := make([][]*dto.StepDefinition, len(opts.Pickle.Steps))
	// for i := range opts.Pickle.Steps {
	// 	// TODO lookup matching step definitions
	// 	stepIndexToStepDefinitions[i] = []*dto.StepDefinition{}
	// }
	// TODO filter hook defs by tags
	return &TestCaseRunner{
		pickle:                        opts.Pickle,
		uri:                           opts.URI,
		sendCommand:                   opts.SendCommand,
		sendCommandAndAwaitResponse:   opts.SendCommandAndAwaitResponse,
		beforeTestCaseHookDefinitions: opts.BeforeTestCaseHookDefinitions,
		afterTestCaseHookDefinitions:  opts.AfterTestCaseHookDefinitions,
		stepIndexToStepDefinitions:    stepIndexToStepDefinitions,
		result: &dto.TestResult{
			Duration: 0,
			Status:   dto.StatusPassed,
		},
	}
}

// Run runs a test case
func (t *TestCaseRunner) Run() *dto.TestResult {
	t.sendTestCasePreparedEvent()
	t.sendTestCaseStartedEvent()
	t.sendCommandAndAwaitResponse(&dto.Command{Type: dto.CommandTypeInitializeTestCase})
	for index, runHookOrStepFunc := range t.getRunHookAndStepFuncs() {
		t.sendTestStepStartedEvent(index)
		hookOrStepResult := runHookOrStepFunc()
		t.sendTestStepFinishedEvent(index, hookOrStepResult)
		t.updateResult(hookOrStepResult)
	}
	t.sendTestCaseFinishedEvent()
	return t.result
}

func (t *TestCaseRunner) updateResult(hookOrStepResult *dto.TestResult) {
	t.result.Duration += hookOrStepResult.Duration
	if t.shouldUpdateResultStatus(hookOrStepResult) {
		t.result.Status = hookOrStepResult.Status
	}
	if hookOrStepResult.Exception != "" {
		t.result.Exception = hookOrStepResult.Exception
	}
}

func (t *TestCaseRunner) shouldUpdateResultStatus(hookOrStepResult *dto.TestResult) bool {
	switch hookOrStepResult.Status {
	case dto.StatusFailed, dto.StatusAmbiguous:
		return t.result.Status != dto.StatusFailed || t.result.Status != dto.StatusAmbiguous
	default:
		return t.result.Status == dto.StatusPassed || t.result.Status == dto.StatusSkipped
	}
}

func (t *TestCaseRunner) sendTestStepStartedEvent(index int) {
	t.sendCommand(&dto.Command{
		Type: "event",
		Event: event.NewTestStepStarted(event.NewTestStepStartedOptions{
			Index:  index,
			Pickle: t.pickle,
			URI:    t.uri,
		}),
	})
}

func (t *TestCaseRunner) sendTestStepFinishedEvent(index int, result *dto.TestResult) {
	t.sendCommand(&dto.Command{
		Type: "event",
		Event: event.NewTestStepFinished(event.NewTestStepFinishedOptions{
			Index:  index,
			Pickle: t.pickle,
			Result: result,
			URI:    t.uri,
		}),
	})
}

func (t *TestCaseRunner) sendTestCaseFinishedEvent() {
	t.sendCommand(&dto.Command{
		Type: "event",
		Event: event.NewTestCaseFinished(event.NewTestCaseFinishedOptions{
			Pickle: t.pickle,
			Result: t.result,
			URI:    t.uri,
		}),
	})
}

func (t *TestCaseRunner) sendTestCasePreparedEvent() {
	t.sendCommand(&dto.Command{
		Type: "event",
		Event: event.NewTestCasePrepared(event.NewTestCasePreparedOptions{
			AfterTestCaseHookDefinitions:  t.afterTestCaseHookDefinitions,
			BeforeTestCaseHookDefinitions: t.beforeTestCaseHookDefinitions,
			Pickle: t.pickle,
			StepIndexToStepDefinitions: t.stepIndexToStepDefinitions,
			URI: t.uri,
		}),
	})
}

func (t *TestCaseRunner) sendTestCaseStartedEvent() {
	t.sendCommand(&dto.Command{
		Type: "event",
		Event: event.NewTestCaseStarted(event.NewTestCaseStartedOptions{
			Pickle: t.pickle,
			URI:    t.uri,
		}),
	})
}

func (t *TestCaseRunner) getRunHookAndStepFuncs() []func() *dto.TestResult {
	var result []func() *dto.TestResult
	for _, beforeTestCaseHook := range t.beforeTestCaseHookDefinitions {
		result = append(result, t.runHookFunc(beforeTestCaseHook, true))
	}
	for index, step := range t.pickle.Steps {
		result = append(result, t.runStepFunc(index, step))
	}
	for _, afterTestCaseHook := range t.afterTestCaseHookDefinitions {
		result = append(result, t.runHookFunc(afterTestCaseHook, false))
	}
	return result
}

func (t *TestCaseRunner) runHookFunc(hook *dto.TestCaseHookDefinition, isBeforeHook bool) func() *dto.TestResult {
	// TODO don't run a before hook if the test case result status is skipped
	return func() *dto.TestResult {
		response := t.sendCommandAndAwaitResponse(&dto.Command{
			Type:           dto.CommandTypeRunBeforeTestCaseHook,
			TestCaseHookID: hook.ID,
		})
		return response.HookOrStepResult
	}
}

func (t *TestCaseRunner) runStepFunc(stepIndex int, step *gherkin.PickleStep) func() *dto.TestResult {
	return func() *dto.TestResult {
		if len(t.stepIndexToStepDefinitions[stepIndex]) == 0 {
			return &dto.TestResult{Status: dto.StatusUndefined}
		}
		// TODO don't run the step if ambiguous
		// TODO don't run the step if test case result status isnt passed
		response := t.sendCommandAndAwaitResponse(&dto.Command{
			Type:     dto.CommandTypeRunTestStep,
			TestStep: step,
		})
		return response.HookOrStepResult
	}
}