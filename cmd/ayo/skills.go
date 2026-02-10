package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/alexcabrera/ayo/internal/agent"
	"github.com/alexcabrera/ayo/internal/builtin"
	"github.com/alexcabrera/ayo/internal/cli"
	"github.com/alexcabrera/ayo/internal/config"
	"github.com/alexcabrera/ayo/internal/paths"
	"github.com/alexcabrera/ayo/internal/skills"
)

// Ensure cli package is used
var _ = cli.Output{}

func newSkillsCmd(cfgPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "skill",
		Short:   "Manage skills",
		Aliases: []string{"skills"},
		Long: `Manage skills that extend agent capabilities.

Skills are instruction sets that teach agents specialized tasks.
They follow the agentskills spec (https://agentskills.org).

Discovery priority (first found wins):
  1. Agent-specific skills (in agent's skills/ directory)
  2. User shared skills (~/.config/ayo/skills/)
  3. Built-in skills (~/.local/share/ayo/skills/)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Default to list
			return listSkillsCmd(cfgPath).RunE(cmd, args)
		},
	}

	cmd.AddCommand(listSkillsCmd(cfgPath))
	cmd.AddCommand(showSkillCmd(cfgPath))
	cmd.AddCommand(validateSkillCmd())
	cmd.AddCommand(createSkillCmd(cfgPath))
	cmd.AddCommand(updateSkillsCmd(cfgPath))

	return cmd
}

func listSkillsCmd(cfgPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all available skills",
		RunE: func(cmd *cobra.Command, args []string) error {
			return withConfig(cfgPath, func(cfg config.Config) error {
				// Install builtins if needed
				if err := builtin.Install(); err != nil {
					return fmt.Errorf("install builtins: %w", err)
				}

				// Build source list for discovery (bypasses include filter)
				var sources []skills.SkillSourceDir
				if cfg.SkillsDir != "" {
					sources = append(sources, skills.SkillSourceDir{
						Path:   cfg.SkillsDir,
						Source: skills.SourceUserShared,
						Label:  "user",
					})
				}
				sources = append(sources, skills.SkillSourceDir{
					Path:   builtin.SkillsInstallDir(),
					Source: skills.SourceBuiltIn,
					Label:  "builtin",
				})

				result := skills.DiscoverWithSources(sources)

				if len(result.Skills) == 0 {
					fmt.Println("No skills found.")
					return nil
				}

				// Color palette
				purple := lipgloss.Color("#a78bfa")
				cyan := lipgloss.Color("#67e8f9")
				muted := lipgloss.Color("#6b7280")
				text := lipgloss.Color("#e5e7eb")
				subtle := lipgloss.Color("#374151")

				// Styles
				headerStyle := lipgloss.NewStyle().Bold(true).Foreground(purple)
				sectionStyle := lipgloss.NewStyle().Foreground(muted).Bold(true)
				iconStyle := lipgloss.NewStyle().Foreground(cyan)
				nameStyle := lipgloss.NewStyle().Foreground(cyan).Bold(true)
				descStyle := lipgloss.NewStyle().Foreground(text)
				countStyle := lipgloss.NewStyle().Foreground(muted)
				dividerStyle := lipgloss.NewStyle().Foreground(subtle)
				emptyStyle := lipgloss.NewStyle().Foreground(muted).Italic(true)

				// Group skills by source
				var userSkills, builtinSkills []skills.Metadata
				for _, s := range result.Skills {
					switch s.Source {
					case skills.SourceUserShared:
						userSkills = append(userSkills, s)
					case skills.SourceBuiltIn:
						builtinSkills = append(builtinSkills, s)
					}
				}

				// JSON output
				if globalOutput.JSON {
					type skillJSON struct {
						Name        string `json:"name"`
						Description string `json:"description,omitempty"`
						Builtin     bool   `json:"builtin"`
					}
					var allSkills []skillJSON
					for _, s := range userSkills {
						allSkills = append(allSkills, skillJSON{Name: s.Name, Description: s.Description, Builtin: false})
					}
					for _, s := range builtinSkills {
						allSkills = append(allSkills, skillJSON{Name: s.Name, Description: s.Description, Builtin: true})
					}
					globalOutput.PrintData(allSkills, "")
					return nil
				}

				// Quiet mode: just list names
				if globalOutput.Quiet {
					for _, s := range userSkills {
						fmt.Println(s.Name)
					}
					for _, s := range builtinSkills {
						fmt.Println(s.Name)
					}
					return nil
				}

				// Render function for a skill
				renderSkill := func(s skills.Metadata) {
					icon := iconStyle.Render("◆")
					name := nameStyle.Render(s.Name)
					fmt.Printf("  %s %s\n", icon, name)

					// Description (truncated, indented)
					desc := s.Description
					if len(desc) > 52 {
						desc = desc[:49] + "..."
					}
					if desc != "" {
						fmt.Printf("    %s\n", descStyle.Render(desc))
					}
				}

				// Header
				fmt.Println()
				fmt.Println(headerStyle.Render("  Skills"))
				fmt.Println(dividerStyle.Render("  " + strings.Repeat("─", 58)))

				// User-defined skills section
				fmt.Println()
				fmt.Printf("  %s\n", sectionStyle.Render("User-defined"))
				if len(userSkills) == 0 {
					fmt.Printf("    %s\n", emptyStyle.Render("No user-defined skills"))
					fmt.Printf("    %s\n", emptyStyle.Render("Create one with: ayo skills create <name> --shared"))
				} else {
					for _, s := range userSkills {
						renderSkill(s)
					}
				}

				// Built-in skills section
				fmt.Println()
				fmt.Printf("  %s\n", sectionStyle.Render("Built-in"))
				if len(builtinSkills) == 0 {
					fmt.Printf("    %s\n", emptyStyle.Render("No built-in skills installed"))
					fmt.Printf("    %s\n", emptyStyle.Render("Run: ayo setup"))
				} else {
					for _, s := range builtinSkills {
						renderSkill(s)
					}
				}

				fmt.Println()
				fmt.Println(dividerStyle.Render("  " + strings.Repeat("─", 58)))
				fmt.Println(countStyle.Render(fmt.Sprintf("  %d skills", len(result.Skills))))
				fmt.Println()

				return nil
			})
		},
	}

	return cmd
}

func showSkillCmd(cfgPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <name>",
		Short: "Show skill details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			return withConfig(cfgPath, func(cfg config.Config) error {
				// Install builtins if needed
				if err := builtin.Install(); err != nil {
					return fmt.Errorf("install builtins: %w", err)
				}

				// Build source list for discovery (bypasses include filter)
				builtinDir := builtin.SkillsInstallDir()
				var sources []skills.SkillSourceDir
				for _, dir := range paths.SkillsDirs() {
					// Skip builtin dir if it appears in SkillsDirs
					if dir == builtinDir {
						continue
					}
					sources = append(sources, skills.SkillSourceDir{
						Path:   dir,
						Source: skills.SourceUserShared,
						Label:  "shared",
					})
				}
				sources = append(sources, skills.SkillSourceDir{
					Path:   builtinDir,
					Source: skills.SourceBuiltIn,
					Label:  "builtin",
				})

				result := skills.DiscoverWithSources(sources)

				// Find the skill
				var meta *skills.Metadata
				for i := range result.Skills {
					if result.Skills[i].Name == name {
						meta = &result.Skills[i]
						break
					}
				}

				if meta == nil {
					return fmt.Errorf("skill not found: %s", name)
				}

				// Load full skill
				skill, err := skills.Load(*meta)
				if err != nil {
					return fmt.Errorf("load skill: %w", err)
				}

				// Styles
				headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99"))
				labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
				valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("250"))

				// Build output
				fmt.Println()
				fmt.Println(headerStyle.Render("📚 " + skill.Metadata.Name))
				fmt.Println(strings.Repeat("─", 60))

				fmt.Printf("%s %s\n", labelStyle.Render("Source:"), valueStyle.Render(skill.Metadata.Source.String()))

				if v := skill.Metadata.Version(); v != "" {
					fmt.Printf("%s %s\n", labelStyle.Render("Version:"), valueStyle.Render(v))
				}
				if a := skill.Metadata.Author(); a != "" {
					fmt.Printf("%s %s\n", labelStyle.Render("Author:"), valueStyle.Render(a))
				}
				if skill.Metadata.License != "" {
					fmt.Printf("%s %s\n", labelStyle.Render("License:"), valueStyle.Render(skill.Metadata.License))
				}
				if skill.Metadata.Compatibility != "" {
					fmt.Printf("%s %s\n", labelStyle.Render("Compatibility:"), valueStyle.Render(skill.Metadata.Compatibility))
				}

				fmt.Println(strings.Repeat("─", 60))

				// Description (wrapped)
				fmt.Println(valueStyle.Render(skill.Metadata.Description))

				fmt.Println(strings.Repeat("─", 60))

				fmt.Printf("%s %s\n", labelStyle.Render("Location:"), valueStyle.Render(skill.Metadata.Path))

				// Optional directories
				var extras []string
				if skill.Metadata.HasScripts {
					extras = append(extras, "scripts/")
				}
				if skill.Metadata.HasRefs {
					extras = append(extras, "references/")
				}
				if skill.Metadata.HasAssets {
					extras = append(extras, "assets/")
				}
				if len(extras) > 0 {
					fmt.Printf("%s %s\n", labelStyle.Render("Includes:"), valueStyle.Render(strings.Join(extras, ", ")))
				}

				fmt.Println()

				return nil
			})
		},
	}

	return cmd
}

func validateSkillCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate <path>",
		Short: "Validate a skill directory",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			skillDir := args[0]

			// Make path absolute
			if !filepath.IsAbs(skillDir) {
				cwd, err := os.Getwd()
				if err != nil {
					return err
				}
				skillDir = filepath.Join(cwd, skillDir)
			}

			errors := skills.Validate(skillDir)

			if len(errors) == 0 {
				successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
				fmt.Println(successStyle.Render("✓ Skill is valid"))
				return nil
			}

			errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
			fmt.Println(errorStyle.Render("✗ Validation errors:"))
			for _, e := range errors {
				fmt.Printf("  • %s\n", e.Error())
			}

			return fmt.Errorf("validation failed with %d errors", len(errors))
		},
	}

	return cmd
}

func createSkillCmd(cfgPath *string) *cobra.Command {
	var (
		local     bool
		agentName string
	)

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new skill from template",
		Long: `Create a new skill from template.

By default, skills are created in the shared skills directory (~/.config/ayo/skills/).
Use --local to create in the current directory (project-local skill).
Use --agent to create an agent-specific skill in the agent's skills/ directory.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			// Validate name format
			if errors := validateSkillName(name); len(errors) > 0 {
				return fmt.Errorf("invalid skill name: %s", errors[0])
			}

			// Validate mutually exclusive flags
			if local && agentName != "" {
				return fmt.Errorf("cannot use both --local and --agent flags")
			}

			return withConfig(cfgPath, func(cfg config.Config) error {
				var skillDir string
				switch {
				case agentName != "":
					// --agent: create in agent's skills directory
					handle := agentName
					if !strings.HasPrefix(handle, "@") {
						handle = "@" + handle
					}

					// Install builtins so we can find builtin agents
					if err := builtin.Install(); err != nil {
						return fmt.Errorf("install builtins: %w", err)
					}

					// Find the agent using the loader (checks all agent dirs)
					ag, err := agent.Load(cfg, handle)
					if err != nil {
						return fmt.Errorf("agent not found: %s", handle)
					}

					skillDir = filepath.Join(ag.Dir, "skills", name)
				case local:
					// --local: create in current directory
					cwd, err := os.Getwd()
					if err != nil {
						return err
					}
					skillDir = filepath.Join(cwd, name)
				default:
					// Default: create in shared skills directory
					skillDir = filepath.Join(cfg.SkillsDir, name)
				}

				// Check if already exists
				if _, err := os.Stat(skillDir); err == nil {
					return fmt.Errorf("skill directory already exists: %s", skillDir)
				}

				// Create directory
				if err := os.MkdirAll(skillDir, 0o755); err != nil {
					return err
				}

				// Write template SKILL.md
				template := fmt.Sprintf(`---
name: %s
description: Brief description of what this skill does and when to use it.
metadata:
  author: your-name
  version: "1.0"
---

# %s

## When to Use

Describe the scenarios when this skill should be activated.

## Instructions

Step-by-step instructions for the agent to follow.

## Examples

Show example interactions.
`, name, strings.ReplaceAll(name, "-", " "))

				skillMD := filepath.Join(skillDir, "SKILL.md")
				if err := os.WriteFile(skillMD, []byte(template), 0o644); err != nil {
					return err
				}

				successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
				fmt.Println(successStyle.Render("✓ Created skill: " + name))
				fmt.Printf("  Location: %s\n", skillDir)
				fmt.Println("  Edit SKILL.md to customize your skill.")

				return nil
			})
		},
	}

	cmd.Flags().BoolVar(&local, "local", false, "create in current directory (project-local)")
	cmd.Flags().StringVarP(&agentName, "agent", "a", "", "create as agent-specific skill")

	return cmd
}

func updateSkillsCmd(cfgPath *string) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update built-in skills to latest version",
		RunE: func(cmd *cobra.Command, args []string) error {
			return withConfig(cfgPath, func(cfg config.Config) error {
				sui := newSetupUI(cmd.OutOrStdout())

				if !force {
					// Check for modified skills
					modified, err := builtin.CheckModifiedSkills()
					if err != nil {
						return fmt.Errorf("check modified skills: %w", err)
					}

					if len(modified) > 0 {
						sui.Warning("The following skills have local modifications:")
						for _, m := range modified {
							sui.Info(fmt.Sprintf("  %s: %v", m.Name, m.ModifiedFiles))
						}
						sui.Blank()
						sui.Info("Use --force to overwrite, or copy modifications to user directory first:")
						sui.Info(fmt.Sprintf("  %s", cfg.SkillsDir))
						return fmt.Errorf("skills have local modifications")
					}
				}

				sui.Step("Updating built-in skills...")
				_, err := builtin.ForceInstall()
				if err != nil {
					return err
				}
				sui.SuccessPath("Updated skills at", builtin.SkillsInstallDir())
				return nil
			})
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "overwrite without checking for modifications")

	return cmd
}

func validateSkillName(name string) []string {
	var errors []string

	if name == "" {
		return []string{"name is required"}
	}

	if len(name) > 64 {
		errors = append(errors, "name exceeds 64 characters")
	}

	if name != strings.ToLower(name) {
		errors = append(errors, "name must be lowercase")
	}

	if strings.HasPrefix(name, "-") || strings.HasSuffix(name, "-") {
		errors = append(errors, "name cannot start or end with a hyphen")
	}

	if strings.Contains(name, "--") {
		errors = append(errors, "name cannot contain consecutive hyphens")
	}

	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
			errors = append(errors, "name can only contain lowercase letters, numbers, and hyphens")
			break
		}
	}

	return errors
}
