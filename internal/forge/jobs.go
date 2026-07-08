package forge

import (
	"context"
	"net/http"
)

// JobCreateOpts contains the options for creating a scheduled job.
type JobCreateOpts struct {
	Command   string `json:"command"`
	Frequency string `json:"frequency"` // default "nightly"
	User      string `json:"user"`      // default "forge"
	Name      string `json:"name,omitempty"`
}

// List returns all scheduled jobs on a server.
func (s *JobsService) List(ctx context.Context, serverID int64) ([]ScheduledJob, error) {
	return listResources[ScheduledJob](s.client, ctx, s.client.orgPath("/servers/%d/scheduled-jobs", serverID))
}

// Get returns a single scheduled job by ID.
func (s *JobsService) Get(ctx context.Context, serverID, jobID int64) (*ScheduledJob, error) {
	return getResource[ScheduledJob](s.client, ctx, s.client.orgPath("/servers/%d/scheduled-jobs/%d", serverID, jobID))
}

// Create creates a new scheduled job on a server.
func (s *JobsService) Create(ctx context.Context, serverID int64, opts JobCreateOpts) (*ScheduledJob, error) {
	return createResource[ScheduledJob](s.client, ctx, s.client.orgPath("/servers/%d/scheduled-jobs", serverID), opts)
}

// Delete removes a scheduled job from a server.
func (s *JobsService) Delete(ctx context.Context, serverID, jobID int64) error {
	return s.client.do(ctx, http.MethodDelete, s.client.orgPath("/servers/%d/scheduled-jobs/%d", serverID, jobID), nil, nil)
}
