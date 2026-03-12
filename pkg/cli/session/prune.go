package session

import (
	"fmt"
	"time"
)

// PruneOptions configures the prune operation
type PruneOptions struct {
	DaysOld   int
	DryRun    bool
	ProjectID string
}

// PruneResult holds the outcome of a prune operation
type PruneResult struct {
	Candidates []SessionMetadata
	Deleted    int
	Failed     int
	Errors     []error
	DryRun     bool
}

// PruneSessions identifies and deletes sessions older than the specified threshold.
// In dry-run mode, lists candidates without deleting.
func PruneSessions(accessToken string, opts *PruneOptions) (*PruneResult, error) {
	result := &PruneResult{DryRun: opts.DryRun}

	// Calculate cutoff date
	cutoff := time.Now().AddDate(0, 0, -opts.DaysOld)

	// Query sessions older than cutoff
	metaClient := NewMetadataClient()
	sessions, err := metaClient.Query(accessToken, &QueryOptions{
		ProjectID: opts.ProjectID,
		EndDate:   &cutoff,
		OrderBy:   "created_at",
		OrderDesc: false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query sessions: %w", err)
	}

	result.Candidates = sessions

	if opts.DryRun || len(sessions) == 0 {
		return result, nil
	}

	// Delete each session (storage + metadata)
	storageClient := NewStorageClient()

	for _, s := range sessions {
		// Delete storage object
		if err := storageClient.Delete(accessToken, s.StoragePath); err != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Errorf("storage delete %s: %w", s.ID, err))
			continue
		}

		// Delete metadata record
		if err := metaClient.Delete(accessToken, s.ID); err != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Errorf("metadata delete %s: %w", s.ID, err))
			continue
		}

		result.Deleted++
	}

	return result, nil
}
