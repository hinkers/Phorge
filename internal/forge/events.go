package forge

import (
	"context"
	"fmt"
	"net/http"
)

// List returns all events for a server.
func (s *EventsService) List(ctx context.Context, serverID int64) ([]Event, error) {
	var resp struct {
		Events []Event `json:"events"`
	}
	path := fmt.Sprintf("/servers/%d/events", serverID)
	err := s.client.do(ctx, http.MethodGet, path, nil, &resp)
	return resp.Events, err
}
