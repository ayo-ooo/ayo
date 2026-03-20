package build

import (
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

func (m *Manager) Cleanup() {
	if m.buildDir != "" {
		os.RemoveAll(m.buildDir)
	}
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

func hasDependency(files map[string]string, dep string) bool {
	for _, content := range files {
		if strings.Contains(content, dep) {
			return true
		}
	}
	return false
}
