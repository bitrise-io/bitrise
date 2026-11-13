package httpfetch

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/bartventer/httpcache"
	_ "github.com/bartventer/httpcache/store/memcache"
)

// slogAdapter is an adapter for httpcache's logger.
type slogAdapter struct{ l Logger }

func (h slogAdapter) Enabled(context.Context, slog.Level) bool { return true }

func (h slogAdapter) Handle(_ context.Context, r slog.Record) error {
	msg := r.Message
	r.Attrs(func(a slog.Attr) bool {
		msg += " " + a.String()
		return true
	})
	h.l.Debugf("httpcache: %s", msg)
	return nil
}

func (h slogAdapter) WithAttrs([]slog.Attr) slog.Handler { return h }
func (h slogAdapter) WithGroup(string) slog.Handler      { return h }

// newCachingTransport wraps upstream with the in-memory response cache.
func newCachingTransport(logger Logger, upstream http.RoundTripper) (rt http.RoundTripper, err error) {
	defer func() {
		// httpcache.NewTransport panics with ErrOpenCache rather than returning an
		// error when the store cannot be opened, so turn that one back into an error.
		r := recover()
		if r == nil {
			return
		}
		if rerr, ok := r.(error); ok && errors.Is(rerr, httpcache.ErrOpenCache) {
			rt, err = nil, fmt.Errorf("open in-memory HTTP cache: %w", rerr)
			return
		}
		panic(r)
	}()
	return httpcache.NewTransport(
		"memcache://",
		httpcache.WithUpstream(cacheSuccessesOnly{upstream: upstream}),
		httpcache.WithLogger(slog.New(slogAdapter{l: logger})),
	), nil
}

// cacheSuccessesOnly stops the cache storing anything but a 2xx, like a temporary status 502
type cacheSuccessesOnly struct{ upstream http.RoundTripper }

func (c cacheSuccessesOnly) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := c.upstream.RoundTrip(req)
	if err != nil || resp == nil {
		return resp, err
	}
	// 304 is how revalidation succeeds, forward it
	if resp.StatusCode == http.StatusNotModified {
		return resp, nil
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		resp.Header.Set("Cache-Control", "no-store")
	}
	return resp, nil
}
