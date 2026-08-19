package alias

import (
	"testing"

	"github.com/bitrise-io/bitrise/v2/toolprovider/provider"
	"github.com/stretchr/testify/assert"
)

func TestGetCanonicalToolID(t *testing.T) {
	assert.Equal(t, provider.ToolID("golang"), GetCanonicalToolID("go"))
	assert.Equal(t, provider.ToolID("nodejs"), GetCanonicalToolID("node"))
	// Canonical names and unknown tools are returned unchanged.
	assert.Equal(t, provider.ToolID("golang"), GetCanonicalToolID("golang"))
	assert.Equal(t, provider.ToolID("ruby"), GetCanonicalToolID("ruby"))
	// Garbage/unknown input passes through untouched (no validation here).
	assert.Equal(t, provider.ToolID("garbage-xyz"), GetCanonicalToolID("garbage-xyz"))
}

func TestAliasesFor(t *testing.T) {
	assert.Equal(t, []provider.ToolID{"go"}, AliasesFor("golang"))
	assert.Equal(t, []provider.ToolID{"node"}, AliasesFor("nodejs"))
	// Tools without aliases return nil.
	assert.Nil(t, AliasesFor("ruby"))
	// Passing an alias (not the canonical) yields nothing.
	assert.Nil(t, AliasesFor("go"))
	// Garbage/unknown input yields nothing.
	assert.Nil(t, AliasesFor("garbage-xyz"))
}
