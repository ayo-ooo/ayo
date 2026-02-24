// Package triggers provides a pluggable trigger system for ayo.
// Triggers are used to fire events that invoke agents based on various conditions
// like cron schedules, file changes, webhooks, and custom plugin-provided triggers.
package triggers

import (
	"context"
	"time"
)

// TriggerCategory represents the category of a trigger plugin.
type TriggerCategory string

const (
	// TriggerCategoryPoll is for triggers that poll periodically.
	TriggerCategoryPoll TriggerCategory = "poll"
	// TriggerCategoryPush is for event-driven triggers.
	TriggerCategoryPush TriggerCategory = "push"
	// TriggerCategoryWatch is for file system watching triggers.
	TriggerCategoryWatch TriggerCategory = "watch"
)

// TriggerPlugin is the interface that trigger plugins must implement.
// Plugins can be built-in (like cron, watch) or external (loaded from plugins).
type TriggerPlugin interface {
	// Name returns the unique name of this trigger plugin (e.g., "cron", "imap").
	Name() string

	// Category returns the category of this trigger (poll, push, watch).
	Category() TriggerCategory

	// Description returns a brief description of what this trigger does.
	Description() string

	// ConfigSchema returns the JSON Schema for this trigger's configuration.
	// Returns nil if no schema validation is needed.
	ConfigSchema() map[string]any

	// Init initializes the trigger with the given configuration.
	// Called once when the trigger is registered.
	Init(ctx context.Context, config map[string]any) error

	// Start starts the trigger. The callback is called when the trigger fires.
	// This should return quickly; long-running operations should be in goroutines.
	Start(ctx context.Context, callback EventCallback) error

	// Stop stops the trigger and cleans up resources.
	Stop() error

	// Status returns the current status of the trigger.
	Status() TriggerStatus
}

// EventCallback is the function called when a trigger fires.
type EventCallback func(event TriggerEvent) error

// TriggerEvent represents an event fired by a trigger.
type TriggerEvent struct {
	// TriggerName is the name of the trigger that fired.
	TriggerName string `json:"trigger_name"`

	// TriggerType is the type of trigger (e.g., "cron", "watch", "imap").
	TriggerType string `json:"trigger_type"`

	// Payload contains trigger-specific event data.
	Payload map[string]any `json:"payload,omitempty"`

	// Timestamp is when the event occurred.
	Timestamp time.Time `json:"timestamp"`
}

// TriggerStatus represents the current status of a trigger.
type TriggerStatus struct {
	// Running indicates whether the trigger is currently active.
	Running bool `json:"running"`

	// LastEvent is the timestamp of the last event fired.
	LastEvent time.Time `json:"last_event,omitempty"`

	// EventCount is the total number of events fired.
	EventCount int64 `json:"event_count"`

	// ErrorCount is the number of errors encountered.
	ErrorCount int64 `json:"error_count"`

	// LastError is the most recent error message, if any.
	LastError string `json:"last_error,omitempty"`
}

// TriggerFactory creates a new instance of a trigger plugin.
type TriggerFactory func() TriggerPlugin

// TriggerInfo contains metadata about a registered trigger type.
type TriggerInfo struct {
	// Name is the trigger type name.
	Name string `json:"name"`

	// Category is the trigger category.
	Category TriggerCategory `json:"category"`

	// Description describes what this trigger does.
	Description string `json:"description"`

	// PluginName is the name of the plugin providing this trigger.
	// Empty string for built-in triggers.
	PluginName string `json:"plugin_name,omitempty"`

	// ConfigSchema is the JSON Schema for configuration.
	ConfigSchema map[string]any `json:"config_schema,omitempty"`
}
