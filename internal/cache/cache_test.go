package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew_IsEmpty(t *testing.T) {
	c := New()
	_, ok := c.LookupApp("anything")
	assert.False(t, ok, "fresh cache should be empty")
}

func TestSetAndLookupApp(t *testing.T) {
	c := New()
	c.SetApp("My App", "abc12345")

	slug, ok := c.LookupApp("My App")
	assert.True(t, ok)
	assert.Equal(t, "abc12345", slug)

	slug, ok = c.LookupApp("my app")
	assert.True(t, ok, "lookup should be case-insensitive")
	assert.Equal(t, "abc12345", slug)
}

func TestNilCacheIsNoop(t *testing.T) {
	var c *Cache
	c.SetApp("x", "y")

	_, ok := c.LookupApp("x")
	assert.False(t, ok, "nil cache LookupApp should return false")
}

func TestNew_WorkspacesIsEmpty(t *testing.T) {
	c := New()
	_, ok := c.LookupWorkspace("anything")
	assert.False(t, ok, "fresh cache should be empty")
}

func TestSetAndLookupWorkspace(t *testing.T) {
	c := New()
	c.SetWorkspace("My Workspace", "abc12345")

	slug, ok := c.LookupWorkspace("My Workspace")
	assert.True(t, ok)
	assert.Equal(t, "abc12345", slug)

	slug, ok = c.LookupWorkspace("my workspace")
	assert.True(t, ok, "lookup should be case-insensitive")
	assert.Equal(t, "abc12345", slug)
}

func TestNilCacheWorkspaceIsNoop(t *testing.T) {
	var c *Cache
	c.SetWorkspace("x", "y")

	_, ok := c.LookupWorkspace("x")
	assert.False(t, ok, "nil cache LookupWorkspace should return false")
}
