package generate

import (
	"strings"
	"testing"

	"github.com/charmbracelet/ayo/internal/project"
)

func TestGenerateHooks_HookRunnerStruct(t *testing.T) {
	proj := &project.Project{
		Path: "/test/agent",
		Config: project.AgentConfig{
			Name: "test-agent",
		},
	}

	code, err := GenerateHooks(proj, "main")
	if err != nil {
		t.Fatalf("GenerateHooks() error = %v", err)
	}

	if !strings.Contains(code, "type HookRunner struct {") {
		t.Error("Missing HookRunner struct")
	}

	if !strings.Contains(code, "embeddedHooks map[HookType][]byte") {
		t.Error("HookRunner should have embeddedHooks field")
	}

	if !strings.Contains(code, "userHooks     map[HookType]string") {
		t.Error("HookRunner should have userHooks field")
	}

	if !strings.Contains(code, "tempDir       string") {
		t.Error("HookRunner should have tempDir field")
	}
}

func TestGenerateHooks_HookTypes(t *testing.T) {
	proj := &project.Project{
		Path: "/test/agent",
		Config: project.AgentConfig{
			Name: "test-agent",
		},
	}

	code, err := GenerateHooks(proj, "main")
	if err != nil {
		t.Fatalf("GenerateHooks() error = %v", err)
	}

	hookTypes := []string{
		"HookAgentStart",
		"HookAgentFinish",
		"HookAgentError",
		"HookTextStart",
		"HookTextDelta",
		"HookTextEnd",
		"HookToolCall",
		"HookToolResult",
	}

	for _, hookType := range hookTypes {
		if !strings.Contains(code, hookType) {
			t.Errorf("Missing hook type: %s", hookType)
		}
	}
}

func TestGenerateHooks_ExecutionOrder(t *testing.T) {
	proj := &project.Project{
		Path: "/test/agent",
		Config: project.AgentConfig{
			Name: "test-agent",
		},
	}

	code, err := GenerateHooks(proj, "main")
	if err != nil {
		t.Fatalf("GenerateHooks() error = %v", err)
	}

	// Find the Run method
	runIdx := strings.Index(code, "func (r *HookRunner) Run(")
	if runIdx == -1 {
		t.Fatal("Missing Run method")
	}

	embeddedIdx := strings.Index(code[runIdx:], "r.embeddedHooks")
	userIdx := strings.Index(code[runIdx:], "r.userHooks")

	if embeddedIdx == -1 {
		t.Error("Run method should check embedded hooks")
	}
	if userIdx == -1 {
		t.Error("Run method should check user hooks")
	}

	// Embedded should come before user in the code
	if embeddedIdx > userIdx {
		t.Error("Embedded hooks should run BEFORE user hooks")
	}
}

func TestGenerateHooks_NonBlockingBehavior(t *testing.T) {
	proj := &project.Project{
		Path: "/test/agent",
		Config: project.AgentConfig{
			Name: "test-agent",
		},
	}

	code, err := GenerateHooks(proj, "main")
	if err != nil {
		t.Fatalf("GenerateHooks() error = %v", err)
	}

	// Check that hook failures are logged but don't block
	if !strings.Contains(code, "log.Printf") {
		t.Error("Hook failures should be logged")
	}

	// The Run method ends with return nil (non-blocking)
	if !strings.Contains(code, "return nil\n}\n\n") {
		t.Error("Run method should return nil (non-blocking)")
	}
}

func TestGenerateHooks_PayloadSerialization(t *testing.T) {
	proj := &project.Project{
		Path: "/test/agent",
		Config: project.AgentConfig{
			Name: "test-agent",
		},
	}

	code, err := GenerateHooks(proj, "main")
	if err != nil {
		t.Fatalf("GenerateHooks() error = %v", err)
	}

	if !strings.Contains(code, "serializeHookPayload") {
		t.Error("Missing serializeHookPayload function")
	}

	if !strings.Contains(code, `"event"`) {
		t.Error("Payload should include event field")
	}

	if !strings.Contains(code, `"timestamp"`) {
		t.Error("Payload should include timestamp field")
	}

	if !strings.Contains(code, `"data"`) {
		t.Error("Payload should include data field")
	}

	if !strings.Contains(code, "json.Marshal") {
		t.Error("Payload should be JSON marshaled")
	}
}

func TestGenerateHooks_TimestampFormat(t *testing.T) {
	proj := &project.Project{
		Path: "/test/agent",
		Config: project.AgentConfig{
			Name: "test-agent",
		},
	}

	code, err := GenerateHooks(proj, "main")
	if err != nil {
		t.Fatalf("GenerateHooks() error = %v", err)
	}

	if !strings.Contains(code, "time.RFC3339") {
		t.Error("Timestamp should use RFC3339 format")
	}
}

func TestGenerateHooks_EmbeddedHookExecution(t *testing.T) {
	proj := &project.Project{
		Path: "/test/agent",
		Config: project.AgentConfig{
			Name: "test-agent",
		},
	}

	code, err := GenerateHooks(proj, "main")
	if err != nil {
		t.Fatalf("GenerateHooks() error = %v", err)
	}

	if !strings.Contains(code, "executeEmbedded") {
		t.Error("Missing executeEmbedded method")
	}

	if !strings.Contains(code, "os.WriteFile(tmpFile, content, 0755)") {
		t.Error("Embedded hooks should be written to temp file with executable permissions")
	}

	if !strings.Contains(code, "defer os.Remove(tmpFile)") {
		t.Error("Temp files should be cleaned up")
	}
}

func TestGenerateHooks_UserHookExecution(t *testing.T) {
	proj := &project.Project{
		Path: "/test/agent",
		Config: project.AgentConfig{
			Name: "test-agent",
		},
	}

	code, err := GenerateHooks(proj, "main")
	if err != nil {
		t.Fatalf("GenerateHooks() error = %v", err)
	}

	if !strings.Contains(code, "executeUser") {
		t.Error("Missing executeUser method")
	}
}

func TestGenerateHooks_ExecuteHookUsesContext(t *testing.T) {
	proj := &project.Project{
		Path: "/test/agent",
		Config: project.AgentConfig{
			Name: "test-agent",
		},
	}

	code, err := GenerateHooks(proj, "main")
	if err != nil {
		t.Fatalf("GenerateHooks() error = %v", err)
	}

	if !strings.Contains(code, "exec.CommandContext(ctx") {
		t.Error("Hook execution should use context for cancellation")
	}
}

func TestGenerateHooks_PayloadViaStdin(t *testing.T) {
	proj := &project.Project{
		Path: "/test/agent",
		Config: project.AgentConfig{
			Name: "test-agent",
		},
	}

	code, err := GenerateHooks(proj, "main")
	if err != nil {
		t.Fatalf("GenerateHooks() error = %v", err)
	}

	if !strings.Contains(code, "cmd.Stdin = bytes.NewReader(payload)") {
		t.Error("Hook payload should be passed via stdin")
	}
}

func TestGenerateHooks_NewHookRunner(t *testing.T) {
	proj := &project.Project{
		Path: "/test/agent",
		Config: project.AgentConfig{
			Name: "test-agent",
		},
	}

	code, err := GenerateHooks(proj, "main")
	if err != nil {
		t.Fatalf("GenerateHooks() error = %v", err)
	}

	if !strings.Contains(code, "func NewHookRunner(") {
		t.Error("Missing NewHookRunner constructor")
	}
}

func TestGenerateHooks_InitHooks(t *testing.T) {
	proj := &project.Project{
		Path: "/test/agent",
		Config: project.AgentConfig{
			Name: "test-agent",
		},
		Hooks: map[project.HookType]string{
			project.HookAgentStart: "hooks/agent-start",
		},
	}

	code, err := GenerateHooks(proj, "main")
	if err != nil {
		t.Fatalf("GenerateHooks() error = %v", err)
	}

	if !strings.Contains(code, "func initHooks()") {
		t.Error("Missing initHooks function")
	}

	if !strings.Contains(code, "os.MkdirTemp") {
		t.Error("initHooks should create temp directory")
	}
}

func TestGenerateHooks_EmbeddedHooksWired(t *testing.T) {
	proj := &project.Project{
		Path: "/test/agent",
		Config: project.AgentConfig{
			Name: "test-agent",
		},
		Hooks: map[project.HookType]string{
			project.HookAgentStart:  "hooks/agent-start",
			project.HookAgentFinish: "hooks/agent-finish",
		},
	}

	code, err := GenerateHooks(proj, "main")
	if err != nil {
		t.Fatalf("GenerateHooks() error = %v", err)
	}

	// Should wire up the embedded hooks using the hook variables
	if !strings.Contains(code, "embedded[HookAgentStart] = hookAgentStart") {
		t.Error("Should wire embedded agent-start hook")
	}

	if !strings.Contains(code, "embedded[HookAgentFinish] = hookAgentFinish") {
		t.Error("Should wire embedded agent-finish hook")
	}
}
