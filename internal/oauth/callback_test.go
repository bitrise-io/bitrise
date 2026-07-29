package oauth

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCallbackServer_HappyPath(t *testing.T) {
	cs := startCallbackServer(t, "st8")
	visitCallback(t, cs, "?code=abc&state=st8")

	code, err := cs.wait(waitCtx(t))
	require.NoError(t, err)
	assert.Equal(t, "abc", code)
}

func TestCallbackServer_StateMismatch(t *testing.T) {
	cs := startCallbackServer(t, "right")
	visitCallback(t, cs, "?code=abc&state=wrong")

	_, err := cs.wait(waitCtx(t))
	assert.ErrorContains(t, err, "state mismatch")
}

func TestCallbackServer_AuthorizationDenied(t *testing.T) {
	cs := startCallbackServer(t, "st8")
	visitCallback(t, cs, "?error=access_denied&error_description=User+cancelled&state=st8")

	_, err := cs.wait(waitCtx(t))
	assert.ErrorContains(t, err, "authorization denied")
	assert.ErrorContains(t, err, "access_denied")
}

func TestCallbackServer_MissingCode(t *testing.T) {
	cs := startCallbackServer(t, "st8")
	visitCallback(t, cs, "?state=st8")

	_, err := cs.wait(waitCtx(t))
	assert.ErrorContains(t, err, "missing authorization code")
}

func TestCallbackServer_Timeout(t *testing.T) {
	cs := startCallbackServer(t, "s")

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := cs.wait(ctx)
	assert.Error(t, err, "expected a timeout error when no callback arrives")
}

func startCallbackServer(t *testing.T, state string) *callbackServer {
	t.Helper()
	cs, err := newCallbackServer(state)
	require.NoError(t, err)
	t.Cleanup(cs.close)
	cs.start()
	return cs
}

// visitCallback hits the loopback callback the way the browser would. It can
// run synchronously ahead of wait(): the handler delivers into a buffered
// channel, so it never blocks on a reader.
func visitCallback(t *testing.T, cs *callbackServer, query string) {
	t.Helper()
	resp, err := http.Get(cs.redirectURI() + query)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
}

func waitCtx(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	t.Cleanup(cancel)
	return ctx
}
