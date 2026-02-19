package goals

import (
	"context"
	"encoding/json"
	"testing"
)

func TestNewGoalsTool(t *testing.T) {
	p := &Plugin{}
	tool := p.newGoalsTool()

	info := tool.Info()
	if info.Name != ToolName {
		t.Errorf("tool name = %q, want %q", info.Name, ToolName)
	}
	if info.Description != ToolDescription {
		t.Errorf("tool description = %q, want %q", info.Description, ToolDescription)
	}
}

func TestHandleGoals_NotInitialized(t *testing.T) {
	p := &Plugin{}

	params := GoalsParams{
		Goals: []GoalParam{
			{Goal: "Complete the feature", Status: "active", Progress: 50},
		},
	}

	resp, err := p.handleGoals(context.Background(), params)
	if err != nil {
		t.Fatalf("handleGoals() failed: %v", err)
	}

	// Should return error response since state is nil
	if !resp.IsError {
		t.Error("expected error response for uninitialized plugin")
	}
}

func TestHandleGoals_InvalidStatus(t *testing.T) {
	tmpDir := t.TempDir()
	p := &Plugin{stateDir: tmpDir}
	if err := p.Init(context.Background()); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	params := GoalsParams{
		Goals: []GoalParam{
			{Goal: "Goal 1", Status: "invalid_status", Progress: 50},
		},
	}

	resp, err := p.handleGoals(context.Background(), params)
	if err != nil {
		t.Fatalf("handleGoals() failed: %v", err)
	}

	if !resp.IsError {
		t.Error("expected error response for invalid status")
	}
}

func TestHandleGoals_InvalidProgress(t *testing.T) {
	tmpDir := t.TempDir()
	p := &Plugin{stateDir: tmpDir}
	if err := p.Init(context.Background()); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	params := GoalsParams{
		Goals: []GoalParam{
			{Goal: "Goal 1", Status: "active", Progress: 150}, // Invalid: > 100
		},
	}

	resp, err := p.handleGoals(context.Background(), params)
	if err != nil {
		t.Fatalf("handleGoals() failed: %v", err)
	}

	if !resp.IsError {
		t.Error("expected error response for invalid progress")
	}
}

func TestHandleGoals_Success(t *testing.T) {
	tmpDir := t.TempDir()
	p := &Plugin{stateDir: tmpDir}
	if err := p.Init(context.Background()); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	params := GoalsParams{
		Goals: []GoalParam{
			{
				Goal:     "Complete the feature implementation",
				Status:   "active",
				Progress: 60,
				Milestones: []MilestoneParam{
					{Description: "Design complete", Completed: true},
					{Description: "Tests written", Completed: false},
				},
				Notes: []string{"Started implementation", "Found a good approach"},
			},
			{
				Goal:     "Previous task",
				Status:   "achieved",
				Progress: 100,
			},
		},
	}

	resp, err := p.handleGoals(context.Background(), params)
	if err != nil {
		t.Fatalf("handleGoals() failed: %v", err)
	}

	if resp.IsError {
		t.Errorf("unexpected error response: %s", resp.Content)
	}

	// Parse response
	var result GoalsResult
	if err := json.Unmarshal([]byte(resp.Content), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if result.Active != 1 {
		t.Errorf("Active = %d, want 1", result.Active)
	}
	if result.Achieved != 1 {
		t.Errorf("Achieved = %d, want 1", result.Achieved)
	}
	if result.CurrentGoal != "Complete the feature implementation" {
		t.Errorf("CurrentGoal = %q, want %q", result.CurrentGoal, "Complete the feature implementation")
	}
	if result.CurrentProgress != 60 {
		t.Errorf("CurrentProgress = %d, want 60", result.CurrentProgress)
	}

	// Verify state was updated
	goals := p.state.List()
	if len(goals) != 2 {
		t.Fatalf("state has %d goals, want 2", len(goals))
	}
	if goals[0].Goal != "Complete the feature implementation" {
		t.Errorf("goals[0].Goal = %q, want %q", goals[0].Goal, "Complete the feature implementation")
	}
	if len(goals[0].Milestones) != 2 {
		t.Errorf("goals[0] has %d milestones, want 2", len(goals[0].Milestones))
	}
}

func TestHandleGoals_EmptyList(t *testing.T) {
	tmpDir := t.TempDir()
	p := &Plugin{stateDir: tmpDir}
	if err := p.Init(context.Background()); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	// First add a goal
	params := GoalsParams{
		Goals: []GoalParam{
			{Goal: "Goal 1", Status: "active", Progress: 50},
		},
	}
	_, _ = p.handleGoals(context.Background(), params)

	// Then clear them
	params = GoalsParams{Goals: []GoalParam{}}
	resp, err := p.handleGoals(context.Background(), params)
	if err != nil {
		t.Fatalf("handleGoals() failed: %v", err)
	}

	if resp.IsError {
		t.Errorf("unexpected error response: %s", resp.Content)
	}

	// Verify state was cleared
	if !p.state.IsEmpty() {
		t.Errorf("state should be empty, got %d goals", p.state.Count())
	}
}

func TestHandleGoals_ProgressUpdate(t *testing.T) {
	tmpDir := t.TempDir()
	p := &Plugin{stateDir: tmpDir}
	if err := p.Init(context.Background()); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	// Set initial goal
	params := GoalsParams{
		Goals: []GoalParam{
			{Goal: "Complete the feature", Status: "active", Progress: 25},
		},
	}
	_, _ = p.handleGoals(context.Background(), params)

	// Get initial created time
	initial := p.state.List()[0]
	initialCreatedAt := initial.CreatedAt

	// Update progress
	params = GoalsParams{
		Goals: []GoalParam{
			{Goal: "Complete the feature", Status: "active", Progress: 75},
		},
	}
	resp, err := p.handleGoals(context.Background(), params)
	if err != nil {
		t.Fatalf("handleGoals() failed: %v", err)
	}

	if resp.IsError {
		t.Errorf("unexpected error response: %s", resp.Content)
	}

	// Verify progress was updated but created_at preserved
	updated := p.state.List()[0]
	if updated.Progress != 75 {
		t.Errorf("progress = %d, want 75", updated.Progress)
	}
	if !updated.CreatedAt.Equal(initialCreatedAt) {
		t.Errorf("CreatedAt changed from %v to %v", initialCreatedAt, updated.CreatedAt)
	}
}
