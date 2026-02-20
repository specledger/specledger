//go:build ignore
package workflows

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

// BatchProcessWorkflow orchestrates the ETL pipeline
func BatchProcessWorkflow(ctx workflow.Context, input BatchInput) (*BatchResult, error) {
	options := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, options)

	// Extract
	var extractResult ExtractResult
	if err := workflow.ExecuteActivity(ctx, ExtractActivity, input).Get(ctx, &extractResult); err != nil {
		return nil, err
	}

	// Transform
	var transformResult TransformResult
	if err := workflow.ExecuteActivity(ctx, TransformActivity, extractResult).Get(ctx, &transformResult); err != nil {
		return nil, err
	}

	// Load
	var loadResult LoadResult
	if err := workflow.ExecuteActivity(ctx, LoadActivity, transformResult).Get(ctx, &loadResult); err != nil {
		return nil, err
	}

	return &BatchResult{
		RecordsProcessed: loadResult.RecordsLoaded,
	}, nil
}

type BatchInput struct {
	Source string
}

type BatchResult struct {
	RecordsProcessed int
}

type ExtractResult struct {
	Records []map[string]interface{}
}

type TransformResult struct {
	Records []map[string]interface{}
}

type LoadResult struct {
	RecordsLoaded int
}
