package build

import (
	"fmt"
	"sort"
	"strings"

	"github.com/alexcabrera/ayo/internal/build/types"
	"github.com/spf13/cobra"
)

// GenerateCLI generates a Cobra command from the CLI configuration
func GenerateCLI(config *types.Config, executeFunc func(cmd *cobra.Command, args []string) error) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   config.Agent.Name,
		Short: config.Agent.Description,
		Long:  config.Agent.Description,
		RunE:  executeFunc,
	}

	// Parse CLI mode
	mode := config.CLI.Mode
	if mode == "" {
		mode = "hybrid" // Default mode
	}

	// Build flags from configuration
	flagsByName := make(map[string]types.CLIFlag)
	positionalFlags := make(map[int]types.CLIFlag)

	// Sort flags by name for consistent ordering
	flagNames := make([]string, 0, len(config.CLI.Flags))
	for name := range config.CLI.Flags {
		flagNames = append(flagNames, name)
	}
	sort.Strings(flagNames)

	// Process each flag
	for _, name := range flagNames {
		flag := config.CLI.Flags[name]

		// Store in both maps
		flagsByName[name] = flag
		if flag.Position >= 0 {
			positionalFlags[flag.Position] = flag
		}

		// Generate flag based on type
		var flagVar interface{}

		switch flag.Type {
		case "string":
			var str string
			flagVar = &str
			cmd.Flags().StringVar(flagVar.(*string), flag.Name, getStringDefault(flag.Default), flag.Description)
		case "int":
			var i int
			flagVar = &i
			cmd.Flags().IntVar(flagVar.(*int), flag.Name, getIntDefault(flag.Default), flag.Description)
		case "float":
			var f float64
			flagVar = &f
			cmd.Flags().Float64Var(flagVar.(*float64), flag.Name, getFloatDefault(flag.Default), flag.Description)
		case "bool":
			var b bool
			flagVar = &b
			cmd.Flags().BoolVar(flagVar.(*bool), flag.Name, getBoolDefault(flag.Default), flag.Description)
		case "array":
			var arr []string
			flagVar = &arr
			cmd.Flags().StringArrayVar(flagVar.(*[]string), flag.Name, getArrayDefault(flag.Default), flag.Description)
		default:
			return nil, fmt.Errorf("unsupported flag type: %s", flag.Type)
		}

		// Add short flag if provided
		if flag.Short != "" {
			cmd.Flags().StringVarP(flagVar.(*string), flag.Name, flag.Short, "", flag.Description)
		}

		// Mark as required if specified
		if flag.Required && flag.Position < 0 {
			_ = cmd.MarkFlagRequired(flag.Name)
		}
	}

	// Configure argument handling based on mode
	switch mode {
	case "structured":
		// Structured mode: use positional args for flags with position >= 0
		args := []string{}
		for i := 0; i < len(positionalFlags); i++ {
			if flag, ok := positionalFlags[i]; ok {
				args = append(args, fmt.Sprintf("<%s>", flag.Name))
			}
		}
		cmd.Args = cobra.MinimumNArgs(len(args))
		cmd.Use = fmt.Sprintf("%s %s", config.Agent.Name, strings.Join(args, " "))

	case "freeform":
		// Freeform mode: accept arbitrary arguments as prompt
		cmd.Args = cobra.ArbitraryArgs
		cmd.Use = fmt.Sprintf("%s [prompt...]", config.Agent.Name)

	case "hybrid":
		// Hybrid mode: structured flags + optional freeform prompt
		// Positional args for flags with position >= 0
		maxPos := -1
		for pos := range positionalFlags {
			if pos > maxPos {
				maxPos = pos
			}
		}

		if maxPos >= 0 {
			args := []string{}
			for i := 0; i <= maxPos; i++ {
				if flag, ok := positionalFlags[i]; ok {
					args = append(args, fmt.Sprintf("<%s>", flag.Name))
				}
			}
			cmd.Args = cobra.MinimumNArgs(len(args))
			cmd.Use = fmt.Sprintf("%s %s [prompt...]", config.Agent.Name, strings.Join(args, " "))
		} else {
			cmd.Args = cobra.ArbitraryArgs
			cmd.Use = fmt.Sprintf("%s [prompt...]", config.Agent.Name)
		}
	}

	// Add help for available flags
	for _, flag := range config.CLI.Flags {
		if flag.Position >= 0 {
			// Add to usage line
			continue
		}
	}

	return cmd, nil
}

// ParseArgs parses command line arguments into a map of values
func ParseArgs(cmd *cobra.Command, args []string, config *types.Config) (map[string]interface{}, []string, error) {
	result := make(map[string]interface{})

	// Parse flags
	for name, flag := range config.CLI.Flags {
		if flag.Position >= 0 {
			// This is a positional argument
			continue
		}

		// Get flag value based on type
		switch flag.Type {
		case "string":
			if cmd.Flags().Changed(name) {
				val, _ := cmd.Flags().GetString(name)
				result[name] = val
			}
		case "int":
			if cmd.Flags().Changed(name) {
				val, _ := cmd.Flags().GetInt(name)
				result[name] = val
			}
		case "float":
			if cmd.Flags().Changed(name) {
				val, _ := cmd.Flags().GetFloat64(name)
				result[name] = val
			}
		case "bool":
			if cmd.Flags().Changed(name) {
				val, _ := cmd.Flags().GetBool(name)
				result[name] = val
			}
		case "array":
			if cmd.Flags().Changed(name) {
				val, _ := cmd.Flags().GetStringArray(name)
				result[name] = val
			}
		}
	}

	// Parse positional arguments
	positionalArgs := make([]string, 0)
	for i := 0; i < len(args); i++ {
		// Check if this position corresponds to a flag
		for _, flag := range config.CLI.Flags {
			if flag.Position == i {
				// Parse based on flag type
				switch flag.Type {
				case "string":
					result[flag.Name] = args[i]
				case "int":
					var intVal int
					_, err := fmt.Sscanf(args[i], "%d", &intVal)
					if err != nil {
						return nil, nil, fmt.Errorf("invalid integer for flag %s: %s", flag.Name, args[i])
					}
					result[flag.Name] = intVal
				case "float":
					var floatVal float64
					_, err := fmt.Sscanf(args[i], "%f", &floatVal)
					if err != nil {
						return nil, nil, fmt.Errorf("invalid float for flag %s: %s", flag.Name, args[i])
					}
					result[flag.Name] = floatVal
				case "bool":
					result[flag.Name] = strings.ToLower(args[i]) == "true" || args[i] == "1"
				case "array":
					// Split by comma for array positional args
					result[flag.Name] = strings.Split(args[i], ",")
				}
				break
			}
		}
	}

	// Determine which args are positional vs freeform
	positionalCount := 0
	for _, flag := range config.CLI.Flags {
		if flag.Position >= 0 {
			positionalCount++
		}
	}

	// Freeform args are anything beyond the positional args
	if config.CLI.Mode == "freeform" || config.CLI.Mode == "hybrid" {
		if len(args) > positionalCount {
			freeformArgs := args[positionalCount:]
			positionalArgs = freeformArgs
		}
	}

	return result, positionalArgs, nil
}

// Helper functions for default values
func getStringDefault(def any) string {
	if def == nil {
		return ""
	}
	if str, ok := def.(string); ok {
		return str
	}
	return ""
}

func getIntDefault(def any) int {
	if def == nil {
		return 0
	}
	if i, ok := def.(int); ok {
		return i
	}
	if f, ok := def.(float64); ok {
		return int(f)
	}
	return 0
}

func getFloatDefault(def any) float64 {
	if def == nil {
		return 0
	}
	if f, ok := def.(float64); ok {
		return f
	}
	if i, ok := def.(int); ok {
		return float64(i)
	}
	return 0
}

func getBoolDefault(def any) bool {
	if def == nil {
		return false
	}
	if b, ok := def.(bool); ok {
		return b
	}
	return false
}

func getArrayDefault(def any) []string {
	if def == nil {
		return nil
	}
	if arr, ok := def.([]string); ok {
		return arr
	}
	if arr, ok := def.([]interface{}); ok {
		result := make([]string, len(arr))
		for i, v := range arr {
			if str, ok := v.(string); ok {
				result[i] = str
			}
		}
		return result
	}
	return nil
}
