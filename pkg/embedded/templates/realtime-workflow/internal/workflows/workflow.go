//go:build ignore
package workflows

import (
	"time"

	"go.temporal.io/sdk/workflow"

	"example.com/project/internal/activities"
)

// ProcessEventWorkflow handles real-time event processing
func ProcessEventWorkflow(ctx workflow.Context, event Event) (*Result, error) {
	options := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, options)

	// Validate event
	var validated bool
	if err := workflow.ExecuteActivity(ctx, activities.ValidateEvent, event).Get(ctx, &validated); err != nil {
		return nil, err
	}

	if !validated {
		return &Result{Status: "invalid"}, nil
	}

	// Process event
	var processed ProcessedEvent
	if err := workflow.ExecuteActivity(ctx, activities.ProcessEvent, event).Get(ctx, &processed); err != nil {
		return nil, err
	}

	// Notify downstream
	if err := workflow.ExecuteActivity(ctx, activities.NotifyDownstream, processed).Get(ctx, nil); err != nil {
		return nil, err
	}

	return &Result{Status: "completed", EventID: processed.ID}, nil
}

type Event struct {
	ID   string
	Type string
	Data map[string]interface{}
}

type ProcessedEvent struct {
	ID        string
	Processed bool
}

type Result struct {
	Status  string
	EventID string
}
