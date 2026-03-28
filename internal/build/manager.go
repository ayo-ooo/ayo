package build

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ayo-ooo/ayo/internal/generate"
	"github.com/ayo-ooo/ayo/internal/project"
	"github.com/ayo-ooo/ayo/internal/schema"
)

// Platform represents a GOOS/GOARCH pair for cross-compilation.
type Platform struct {
	OS   string
	Arch string
}

// String returns the platform as "os/arch".
func (p Platform) String() string {
	return p.OS + "/" + p.Arch
}

// Suffix returns a string like "-linux-amd64" for use in binary names.
func (p Platform) Suffix() string {
	return "-" + p.OS + "-" + p.Arch
}

// AllPlatforms returns the default set of cross-compilation targets.
func AllPlatforms() []Platform {
	return []Platform{
		{OS: "darwin", Arch: "amd64"},
		{OS: "darwin", Arch: "arm64"},
		{OS: "linux", Arch: "amd64"},
		{OS: "linux", Arch: "arm64"},
	}
}

// ParsePlatform parses a string like "linux/amd64" into a Platform.
func ParsePlatform(s string) (Platform, error) {
	parts := strings.SplitN(s, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return Platform{}, fmt.Errorf("invalid platform %q: expected format os/arch", s)
	}
	return Platform{OS: parts[0], Arch: parts[1]}, nil
}

type Manager struct {
	buildDir string
}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) Build(proj *project.Project, outputPath string) error {
	var err error
	m.buildDir, err = os.MkdirTemp("", "ayo-build-*")
	if err != nil {
		return fmt.Errorf("creating build directory: %w", err)
	}
	defer m.Cleanup()

	if err := m.generateFiles(proj); err != nil {
		return fmt.Errorf("generating files: %w", err)
	}

	if err := m.copyAssets(proj); err != nil {
		return fmt.Errorf("copying assets: %w", err)
	}

	if outputPath == "" {
		outputPath = proj.Config.Name
		if runtime.GOOS == "windows" {
			outputPath += ".exe"
		}
	}

	absOutput, err := filepath.Abs(outputPath)
	if err != nil {
		return fmt.Errorf("resolving output path: %w", err)
	}

	if err := m.compile(absOutput); err != nil {
		return fmt.Errorf("compiling: %w", err)
	}

	return nil
}

// BuildCross compiles the project for each of the given platforms, producing
// one binary per platform with a suffix like "-linux-amd64".
func (m *Manager) BuildCross(proj *project.Project, outputBase string, platforms []Platform) error {
	var err error
	m.buildDir, err = os.MkdirTemp("", "ayo-build-*")
	if err != nil {
		return fmt.Errorf("creating build directory: %w", err)
	}
	defer m.Cleanup()

	if err := m.generateFiles(proj); err != nil {
		return fmt.Errorf("generating files: %w", err)
	}

	if err := m.copyAssets(proj); err != nil {
		return fmt.Errorf("copying assets: %w", err)
	}

	if outputBase == "" {
		outputBase = proj.Config.Name
	}

	// Run go mod tidy once before compiling.
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = m.buildDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go mod tidy failed: %w\n%s", err, string(output))
	}

	for _, plat := range platforms {
		outPath := outputBase + plat.Suffix()
		if plat.OS == "windows" {
			outPath += ".exe"
		}

		absOutput, err := filepath.Abs(outPath)
		if err != nil {
			return fmt.Errorf("resolving output path: %w", err)
		}

		if err := m.compileFor(absOutput, plat); err != nil {
			return fmt.Errorf("compiling for %s: %w", plat, err)
		}
	}

	return nil
}

func (m *Manager) generateFiles(proj *project.Project) error {
	gen := generate.NewGenerator()
	pkgName := "main"

	files := make(map[string]string)

	types, err := generate.GenerateTypes(
		getSchema(proj.Input),
		getSchema(proj.Output),
		pkgName,
	)
	if err != nil {
		return fmt.Errorf("generating types: %w", err)
	}
	files["types.go"] = types

	genFiles, err := gen.Generate(proj, pkgName)
	if err != nil {
		return fmt.Errorf("generating code: %w", err)
	}
	for name, content := range genFiles {
		files[name] = content
	}

	files["go.mod"] = generate.GenerateGoMod(proj)

	for name, content := range files {
		path := filepath.Join(m.buildDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("writing %s: %w", name, err)
		}
	}

	return nil
}

func (m *Manager) copyAssets(proj *project.Project) error {
	systemPath := filepath.Join(m.buildDir, "system.md")
	if err := os.WriteFile(systemPath, []byte(proj.System), 0644); err != nil {
		return fmt.Errorf("copying system.md: %w", err)
	}

	// Generate ayo-metadata.json for self-description protocol
	if err := m.writeMetadata(proj); err != nil {
		return fmt.Errorf("writing metadata: %w", err)
	}

	if proj.Prompt != nil {
		promptPath := filepath.Join(m.buildDir, "prompt.tmpl")
		if err := os.WriteFile(promptPath, []byte(*proj.Prompt), 0644); err != nil {
			return fmt.Errorf("copying prompt.tmpl: %w", err)
		}
	}

	if len(proj.Skills) > 0 {
		skillsDir := filepath.Join(m.buildDir, "skills")
		if err := os.MkdirAll(skillsDir, 0755); err != nil {
			return fmt.Errorf("creating skills directory: %w", err)
		}
		for _, skill := range proj.Skills {
			if err := copyDirectory(skill.Path, filepath.Join(skillsDir, skill.Name)); err != nil {
				return fmt.Errorf("copying skill %s: %w", skill.Name, err)
			}
		}
	}

	if len(proj.Hooks) > 0 {
		hooksDir := filepath.Join(m.buildDir, "hooks")
		if err := os.MkdirAll(hooksDir, 0755); err != nil {
			return fmt.Errorf("creating hooks directory: %w", err)
		}
		for hookType, hookPath := range proj.Hooks {
			data, err := os.ReadFile(hookPath)
			if err != nil {
				return fmt.Errorf("reading hook %s: %w", hookType, err)
			}
			destPath := filepath.Join(hooksDir, string(hookType))
			if err := os.WriteFile(destPath, data, 0755); err != nil {
				return fmt.Errorf("copying hook %s: %w", hookType, err)
			}
		}
	}

	return nil
}

func (m *Manager) compile(outputPath string) error {
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = m.buildDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go mod tidy failed: %w\n%s", err, string(output))
	}

	ldflags := "-s -w"
	cmd = exec.Command("go", "build", "-ldflags", ldflags, "-o", outputPath, ".")
	cmd.Dir = m.buildDir
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go build failed: %w\n%s", err, string(output))
	}

	return nil
}

func (m *Manager) compileFor(outputPath string, plat Platform) error {
	ldflags := "-s -w"
	cmd := exec.Command("go", "build", "-ldflags", ldflags, "-o", outputPath, ".")
	cmd.Dir = m.buildDir
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0",
		"GOOS="+plat.OS,
		"GOARCH="+plat.Arch,
	)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go build failed: %w\n%s", err, string(output))
	}

	return nil
}

func (m *Manager) Cleanup() {
	if m.buildDir != "" {
		os.RemoveAll(m.buildDir)
	}
}

// AgentMetadata is the self-description payload embedded in every ayo agent.
type AgentMetadata struct {
	AyoVersion  int              `json:"ayo_version"`
	Name        string           `json:"name"`
	Version     string           `json:"version"`
	Description string           `json:"description"`
	Type        string           `json:"type"`
	InputSchema json.RawMessage  `json:"input_schema,omitempty"`
	OutputSchema json.RawMessage `json:"output_schema,omitempty"`
	Skills      []string         `json:"skills,omitempty"`
	Hooks       []string         `json:"hooks,omitempty"`
}

func (m *Manager) writeMetadata(proj *project.Project) error {
	agentType := "conversational"
	if proj.Input != nil {
		agentType = "tool"
	}

	meta := AgentMetadata{
		AyoVersion:  1,
		Name:        proj.Config.Name,
		Version:     proj.Config.Version,
		Description: proj.Config.Description,
		Type:        agentType,
	}

	if proj.Input != nil {
		meta.InputSchema = json.RawMessage(proj.Input.Content)
	}
	if proj.Output != nil {
		meta.OutputSchema = json.RawMessage(proj.Output.Content)
	}

	for _, skill := range proj.Skills {
		meta.Skills = append(meta.Skills, skill.Name)
	}
	for hookType := range proj.Hooks {
		meta.Hooks = append(meta.Hooks, string(hookType))
	}

	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling metadata: %w", err)
	}

	metaPath := filepath.Join(m.buildDir, "ayo-metadata.json")
	if err := os.WriteFile(metaPath, data, 0644); err != nil {
		return fmt.Errorf("writing ayo-metadata.json: %w", err)
	}

	return nil
}

func getSchema(s *project.Schema) *schema.ParsedSchema {
	if s == nil {
		return nil
	}
	if parsed, ok := s.Parsed.(*schema.ParsedSchema); ok {
		return parsed
	}
	return nil
}

func copyDirectory(src, dst string) error {
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDirectory(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			data, err := os.ReadFile(srcPath)
			if err != nil {
				return err
			}
			if err := os.WriteFile(dstPath, data, 0644); err != nil {
				return err
			}
		}
	}

	return nil
}

