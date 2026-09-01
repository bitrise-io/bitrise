package rdeapi

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestSessionStartFileUpload_BodyPathAndParse(t *testing.T) {
	rs := newRecordingServer(t, `{"signedUrl":"https://storage/put?sig=1","uploadId":"up-1"}`)

	resp, err := rs.client().SessionStartFileUpload(context.Background(), "ws-1", "s1", StartFileUploadRequest{
		DestinationFolder: "/workspace/repo",
	})
	if err != nil {
		t.Fatalf("SessionStartFileUpload: %v", err)
	}
	if rs.lastMethod != http.MethodPost {
		t.Errorf("method = %s, want POST", rs.lastMethod)
	}
	if want := "/v1/workspaces/ws-1/sessions/s1/start-file-upload"; rs.lastPath != want {
		t.Errorf("path = %s, want %s", rs.lastPath, want)
	}
	var sent StartFileUploadRequest
	if err := json.Unmarshal(rs.lastBody, &sent); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if sent.DestinationFolder != "/workspace/repo" {
		t.Errorf("sent destinationFolder = %q", sent.DestinationFolder)
	}
	if resp.SignedURL == "" || resp.UploadID != "up-1" {
		t.Errorf("resp = %+v", resp)
	}
}

func TestSessionCompleteFileUpload_BodyAndDiscardsResponse(t *testing.T) {
	rs := newRecordingServer(t, `{}`)

	err := rs.client().SessionCompleteFileUpload(context.Background(), "ws-1", "s1", CompleteFileUploadRequest{
		UploadID:          "up-1",
		DestinationFolder: "/workspace/repo",
	})
	if err != nil {
		t.Fatalf("SessionCompleteFileUpload: %v", err)
	}
	if want := "/v1/workspaces/ws-1/sessions/s1/complete-file-upload"; rs.lastPath != want {
		t.Errorf("path = %s, want %s", rs.lastPath, want)
	}
	var sent CompleteFileUploadRequest
	if err := json.Unmarshal(rs.lastBody, &sent); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if sent.UploadID != "up-1" {
		t.Errorf("sent uploadId = %q, want up-1", sent.UploadID)
	}
}

func TestSessionDownloadFile_BodyPathAndParse(t *testing.T) {
	rs := newRecordingServer(t, `{"signedUrl":"https://storage/get?sig=1"}`)

	resp, err := rs.client().SessionDownloadFile(context.Background(), "ws-1", "s1", DownloadFileRequest{
		SourcePath:           "/workspace/repo/dist",
		OnlyContentsOfFolder: true,
	})
	if err != nil {
		t.Fatalf("SessionDownloadFile: %v", err)
	}
	if want := "/v1/workspaces/ws-1/sessions/s1/download-file"; rs.lastPath != want {
		t.Errorf("path = %s, want %s", rs.lastPath, want)
	}
	var sent DownloadFileRequest
	if err := json.Unmarshal(rs.lastBody, &sent); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if sent.SourcePath != "/workspace/repo/dist" || !sent.OnlyContentsOfFolder {
		t.Errorf("sent = %+v", sent)
	}
	if resp.SignedURL == "" {
		t.Error("signed URL is empty")
	}
}

func TestTransfer_ValidationGuards(t *testing.T) {
	rs := newRecordingServer(t, `{}`)
	c := rs.client()
	ctx := context.Background()

	cases := map[string]func() error{
		"StartUpload/no-ws":         func() error { _, err := c.SessionStartFileUpload(ctx, "", "s1", StartFileUploadRequest{}); return err },
		"StartUpload/no-session":    func() error { _, err := c.SessionStartFileUpload(ctx, "ws", "", StartFileUploadRequest{}); return err },
		"CompleteUpload/no-ws":      func() error { return c.SessionCompleteFileUpload(ctx, "", "s1", CompleteFileUploadRequest{}) },
		"CompleteUpload/no-session": func() error { return c.SessionCompleteFileUpload(ctx, "ws", "", CompleteFileUploadRequest{}) },
		"Download/no-ws": func() error {
			_, err := c.SessionDownloadFile(ctx, "", "s1", DownloadFileRequest{SourcePath: "/x"})
			return err
		},
		"Download/no-session": func() error {
			_, err := c.SessionDownloadFile(ctx, "ws", "", DownloadFileRequest{SourcePath: "/x"})
			return err
		},
		"Download/no-source": func() error { _, err := c.SessionDownloadFile(ctx, "ws", "s1", DownloadFileRequest{}); return err },
	}
	for name, call := range cases {
		t.Run(name, func(t *testing.T) {
			if err := call(); err == nil {
				t.Error("expected validation error, got nil")
			}
		})
	}
	if rs.hits != 0 {
		t.Errorf("validation guards made %d HTTP call(s); should short-circuit", rs.hits)
	}
}
