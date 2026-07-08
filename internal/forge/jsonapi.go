package forge

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
)

// resource is the JSON:API "resource object" envelope wrapping a single
// item's attributes.
type resource[T any] struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	Attributes T      `json:"attributes"`
}

// listResponse is the JSON:API envelope for a paginated collection.
type listResponse[T any] struct {
	Data []resource[T] `json:"data"`
	Meta struct {
		PerPage    int     `json:"per_page"`
		NextCursor *string `json:"next_cursor"`
		PrevCursor *string `json:"prev_cursor"`
	} `json:"meta"`
}

// singleResponse is the JSON:API envelope for a single resource.
type singleResponse[T any] struct {
	Data resource[T] `json:"data"`
}

// getResource fetches a single resource and unwraps its attributes.
func getResource[T any](c *Client, ctx context.Context, path string) (*T, error) {
	var resp singleResponse[T]
	if err := c.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	setResourceID(&resp.Data.Attributes, resp.Data.ID)
	return &resp.Data.Attributes, nil
}

// listResources fetches a collection of resources and unwraps their attributes.
func listResources[T any](c *Client, ctx context.Context, path string) ([]T, error) {
	var resp listResponse[T]
	if err := c.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	items := make([]T, len(resp.Data))
	for i, r := range resp.Data {
		items[i] = r.Attributes
		setResourceID(&items[i], r.ID)
	}
	return items, nil
}

// createResource POSTs a new resource and unwraps the created attributes.
func createResource[T any](c *Client, ctx context.Context, path string, body any) (*T, error) {
	var resp singleResponse[T]
	if err := c.do(ctx, http.MethodPost, path, body, &resp); err != nil {
		return nil, err
	}
	setResourceID(&resp.Data.Attributes, resp.Data.ID)
	return &resp.Data.Attributes, nil
}

// updateResource PUTs changes to a resource and unwraps the updated attributes.
func updateResource[T any](c *Client, ctx context.Context, path string, body any) (*T, error) {
	var resp singleResponse[T]
	if err := c.do(ctx, http.MethodPut, path, body, &resp); err != nil {
		return nil, err
	}
	setResourceID(&resp.Data.Attributes, resp.Data.ID)
	return &resp.Data.Attributes, nil
}

// setResourceID copies the JSON:API resource-level ID into the struct's ID
// field. Some resources include id in attributes, others don't — this ensures
// it's always populated.
func setResourceID[T any](result *T, idStr string) {
	if idStr == "" {
		return
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return
	}
	v := reflect.ValueOf(result).Elem()
	f := v.FieldByName("ID")
	if f.IsValid() && f.CanSet() && f.Kind() == reflect.Int64 {
		f.SetInt(id)
	}
}

// orgPath builds an org-scoped API path, e.g. orgPath("/servers/%d", 42)
// produces "/orgs/<org>/servers/42".
func (c *Client) orgPath(format string, args ...any) string {
	allArgs := make([]any, 0, len(args)+1)
	allArgs = append(allArgs, c.org)
	allArgs = append(allArgs, args...)
	return fmt.Sprintf("/orgs/%s"+format, allArgs...)
}
