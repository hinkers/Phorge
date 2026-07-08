package forge

import "context"

// List returns all events for a server.
func (s *EventsService) List(ctx context.Context, serverID int64) ([]Event, error) {
	return listResources[Event](s.client, ctx, s.client.orgPath("/servers/%d/events", serverID))
}
