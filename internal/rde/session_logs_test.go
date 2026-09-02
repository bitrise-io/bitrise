package rde

import (
	"context"
	"testing"
)

func TestLogStageToAPI(t *testing.T) {
	cases := map[string]struct {
		want    string
		wantErr bool
	}{
		"warmup":  {want: "1"},
		"startup": {want: "2"},
		"":        {wantErr: true},
		"main":    {wantErr: true},
	}
	for in, tc := range cases {
		t.Run("stage="+in, func(t *testing.T) {
			got, err := logStageToAPI(in)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("logStageToAPI(%q) = %q, want error", in, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("logStageToAPI(%q): %v", in, err)
			}
			if got != tc.want {
				t.Errorf("logStageToAPI(%q) = %q, want %q", in, got, tc.want)
			}
		})
	}
}

func TestStreamSessionLogs_ForwardsContentAndMapsStage(t *testing.T) {
	rs := newRecordingServer(t, `{"result":{"logContent":"a"}}
{"result":{"heartbeatMessage":true}}
{"result":{"logContent":"b"}}`)

	var got []string
	err := rs.service().StreamSessionLogs(context.Background(), "ws-1", "s1", LogStageStartup, 0, func(s string) error {
		got = append(got, s)
		return nil
	})
	if err != nil {
		t.Fatalf("StreamSessionLogs: %v", err)
	}
	// startup maps to enum 2 in the path.
	if want := "/v1/workspaces/ws-1/sessions/s1/logs/2"; rs.lastPath != want {
		t.Errorf("path = %s, want %s", rs.lastPath, want)
	}
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Errorf("content = %q, want [a b] (heartbeat skipped)", got)
	}
}

func TestStreamSessionLogs_InvalidStageShortCircuits(t *testing.T) {
	rs := newRecordingServer(t, ``)
	err := rs.service().StreamSessionLogs(context.Background(), "ws-1", "s1", "bogus", 0, func(string) error { return nil })
	if err == nil {
		t.Fatal("expected error for invalid stage")
	}
	if rs.lastMethod != "" {
		t.Errorf("invalid stage made an HTTP call (%s); should short-circuit", rs.lastMethod)
	}
}

func TestStreamSessionLogs_NilClient(t *testing.T) {
	if err := NewService(nil).StreamSessionLogs(context.Background(), "ws", "s", LogStageStartup, 0, func(string) error { return nil }); err == nil {
		t.Error("expected error from nil client")
	}
}
