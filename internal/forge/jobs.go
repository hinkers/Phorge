package forge

import (
	"context"
	"fmt"
	"net/http"
)

// JobCreateOpts contains the options for creating a scheduled job.
type JobCreateOpts struct {
	Command   string `json:"command"`
	Frequency string `json:"frequency"` // default "nightly"
	User      string `json:"user"`      // default "forge"
}

// List returns all scheduled jobs on a server.
func (s *JobsService) List(ctx context.Context, serverID int64) ([]ScheduledJob, error) {
	var resp struct {
		Jobs []ScheduledJob `json:"jobs"`
	}
	path := fmt.Sprintf("/servers/%d/jobs", serverID)
	err := s.client.do(ctx, http.MethodGet, path, nil, &resp)
	return resp.Jobs, err
}

// Get returns a single scheduled job by ID.
func (s *JobsService) Get(ctx context.Context, serverID, jobID int64) (*ScheduledJob, error) {
	var resp struct {
		Job ScheduledJob `json:"job"`
	}
	path := fmt.Sprintf("/servers/%d/jobs/%d", serverID, jobID)
	err := s.client.do(ctx, http.MethodGet, path, nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp.Job, nil
}

// Create creates a new scheduled job on a server.
func (s *JobsService) Create(ctx context.Context, serverID int64, opts JobCreateOpts) (*ScheduledJob, error) {
	var resp struct {
		Job ScheduledJob `json:"job"`
	}
	path := fmt.Sprintf("/servers/%d/jobs", serverID)
	err := s.client.do(ctx, http.MethodPost, path, opts, &resp)
	if err != nil {
		return nil, err
	}
	return &resp.Job, nil
}

// Delete removes a scheduled job from a server.
func (s *JobsService) Delete(ctx context.Context, serverID, jobID int64) error {
	path := fmt.Sprintf("/servers/%d/jobs/%d", serverID, jobID)
	return s.client.do(ctx, http.MethodDelete, path, nil, nil)
}
