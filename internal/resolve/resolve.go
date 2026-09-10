// Package resolve maps a user-supplied app or workspace name/slug to the
// canonical slug. Results are stored in a name cache so repeated invocations
// with the same name skip the network call.
package resolve

import (
	"context"
	"fmt"
	"strings"

	"github.com/bitrise-io/bitrise/v2/internal/bitriseapi"
	"github.com/bitrise-io/bitrise/v2/internal/cache"
)

// Resolver maps a display name to an app slug.
type Resolver struct {
	client *bitriseapi.Client
	cache  *cache.Cache
}

// New returns a Resolver backed by the given API client and cache.
// cache may be nil — resolution still works, just without in-memory caching.
func New(client *bitriseapi.Client, c *cache.Cache) *Resolver {
	return &Resolver{client: client, cache: c}
}

// AppSlug maps value to an app slug. See ResolveApp for resolution semantics.
func (r *Resolver) AppSlug(ctx context.Context, value string) (string, error) {
	app, _, err := r.ResolveApp(ctx, value)
	return app.Slug, err
}

// ResolveApp is like AppSlug but returns the full bitriseapi.App when value
// was matched by a name query, so the caller can skip a second GET
// /apps/{slug}. complete=false means value was not resolved by name:
// app.Slug holds the resolved slug (from cache or passthrough) and the
// caller must fetch full data.
func (r *Resolver) ResolveApp(ctx context.Context, value string) (app bitriseapi.App, complete bool, err error) {
	if value == "" {
		return bitriseapi.App{}, false, nil
	}
	if r.cache != nil {
		if slug, ok := r.cache.LookupApp(value); ok {
			return bitriseapi.App{Slug: slug}, false, nil
		}
	}
	apps, _, err := r.client.Apps(ctx, bitriseapi.AppsListOptions{Title: value})
	if err != nil {
		return bitriseapi.App{}, false, fmt.Errorf("look up app %q: %w", value, err)
	}
	var matches []bitriseapi.App
	for _, a := range apps {
		if strings.EqualFold(a.Title, value) {
			matches = append(matches, a)
		}
	}
	switch len(matches) {
	case 0:
		return bitriseapi.App{Slug: value}, false, nil
	case 1:
		if r.cache != nil {
			r.cache.SetApp(matches[0].Title, matches[0].Slug)
		}
		return matches[0], true, nil
	default:
		slugs := make([]string, 0, len(matches))
		for _, m := range matches {
			slugs = append(slugs, m.Slug)
		}
		return bitriseapi.App{}, false, fmt.Errorf("app name %q is ambiguous (matches %d apps: %v) — pass an app ID instead", value, len(matches), slugs)
	}
}

// WorkspaceSlug maps value to a workspace slug. See ResolveWorkspace for
// resolution semantics.
func (r *Resolver) WorkspaceSlug(ctx context.Context, value string) (string, error) {
	org, _, err := r.ResolveWorkspace(ctx, value)
	return org.Slug, err
}

// ResolveWorkspace is like WorkspaceSlug but returns the full
// bitriseapi.Organization when value was matched by a name query, so the
// caller can skip a second lookup. complete=false means value was not
// resolved by name: org.Slug holds the resolved slug (from cache or
// passthrough) and the caller must fetch full data.
//
// Unlike ResolveApp, GET /organizations has no server-side name filter, so
// this always fetches the full workspace list and filters client-side.
func (r *Resolver) ResolveWorkspace(ctx context.Context, value string) (org bitriseapi.Organization, complete bool, err error) {
	if value == "" {
		return bitriseapi.Organization{}, false, nil
	}
	if r.cache != nil {
		if slug, ok := r.cache.LookupWorkspace(value); ok {
			return bitriseapi.Organization{Slug: slug}, false, nil
		}
	}
	orgs, err := r.client.Organizations(ctx)
	if err != nil {
		return bitriseapi.Organization{}, false, fmt.Errorf("look up workspace %q: %w", value, err)
	}
	var matches []bitriseapi.Organization
	for _, o := range orgs {
		if strings.EqualFold(o.Name, value) {
			matches = append(matches, o)
		}
	}
	switch len(matches) {
	case 0:
		return bitriseapi.Organization{Slug: value}, false, nil
	case 1:
		if r.cache != nil {
			r.cache.SetWorkspace(matches[0].Name, matches[0].Slug)
		}
		return matches[0], true, nil
	default:
		slugs := make([]string, 0, len(matches))
		for _, m := range matches {
			slugs = append(slugs, m.Slug)
		}
		return bitriseapi.Organization{}, false, fmt.Errorf("workspace name %q is ambiguous (matches %d workspaces: %v) — pass a workspace ID instead", value, len(matches), slugs)
	}
}
