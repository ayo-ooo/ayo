// Package db provides database functionality (stub for build system)
package db

// FlowRun represents a flow execution record (stub).
type FlowRun struct {
	ID       string
	FlowName string
}

// GetFlowRuns is a stub function.
func GetFlowRuns() ([]FlowRun, error) {
	return []FlowRun{}, nil
}

// SaveFlowRun is a stub function.
func SaveFlowRun(run FlowRun) error {
	return nil
}