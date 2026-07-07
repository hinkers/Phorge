package forge

import "context"

// OrganizationsService provides access to Forge organizations.
type OrganizationsService struct{ client *Client }

// List returns all organizations accessible to the authenticated user.
func (s *OrganizationsService) List(ctx context.Context) ([]Organization, error) {
	return listResources[Organization](s.client, ctx, "/orgs")
}

// Get fetches a single organization by slug.
func (s *OrganizationsService) Get(ctx context.Context, slug string) (*Organization, error) {
	return getResource[Organization](s.client, ctx, "/orgs/"+slug)
}
