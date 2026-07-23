package oauth

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestCallbackServer_HappyPath(t *testing.T) {
	cs, err := newCallbackServer("st8")
	if err != nil {
		t.Fatalf("newCallbackServer: %v", err)
	}
	defer cs.close()
	cs.start()

	go func() {
		resp, err := http.Get(cs.redirectURI() + "?code=abc&state=st8")
		if err == nil {
			_ = resp.Body.Close()
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	code, err := cs.wait(ctx)
	if err != nil {
		t.Fatalf("wait: %v", err)
	}
	if code != "abc" {
		t.Fatalf("code = %q, want abc", code)
	}
}

func TestCallbackServer_StateMismatch(t *testing.T) {
	cs, err := newCallbackServer("right")
	if err != nil {
		t.Fatalf("newCallbackServer: %v", err)
	}
	defer cs.close()
	cs.start()

	go func() {
		resp, err := http.Get(cs.redirectURI() + "?code=abc&state=wrong")
		if err == nil {
			_ = resp.Body.Close()
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err = cs.wait(ctx)
	if err == nil || !strings.Contains(err.Error(), "state mismatch") {
		t.Fatalf("expected state-mismatch error, got %v", err)
	}
}

func TestCallbackServer_Timeout(t *testing.T) {
	cs, err := newCallbackServer("s")
	if err != nil {
		t.Fatalf("newCallbackServer: %v", err)
	}
	defer cs.close()
	cs.start()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if _, err := cs.wait(ctx); err == nil {
		t.Fatal("expected timeout error when no callback arrives")
	}
}
