package rdeapi

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// StartFileUploadRequest is the POST body for /start-file-upload.
type StartFileUploadRequest struct {
	DestinationFolder string `json:"destinationFolder"`
}

// StartFileUploadResponse is the response from /start-file-upload: a signed
// PUT URL plus the upload identifier to send back to complete-file-upload.
type StartFileUploadResponse struct {
	SignedURL string `json:"signedUrl"`
	UploadID  string `json:"uploadId"`
}

// CompleteFileUploadRequest is the POST body for /complete-file-upload.
type CompleteFileUploadRequest struct {
	UploadID          string `json:"uploadId"`
	DestinationFolder string `json:"destinationFolder"`
}

// DownloadFileRequest is the POST body for /download-file.
type DownloadFileRequest struct {
	SourcePath           string `json:"sourcePath"`
	OnlyContentsOfFolder bool   `json:"onlyContentsOfFolder,omitempty"`
}

// DownloadFileResponse carries a signed GET URL for the tar.gz archive.
type DownloadFileResponse struct {
	SignedURL string `json:"signedUrl"`
}

// SessionStartFileUpload initiates an upload to a session, returning a
// signed PUT URL for the tar.gz archive and an upload ID. The actual PUT and
// the tar.gz packing are the caller's responsibility (internal/rde).
// Endpoint: POST /v1/workspaces/{workspaceId}/sessions/{sessionId}/start-file-upload.
func (c *Client) SessionStartFileUpload(ctx context.Context, workspaceID, sessionID string, req StartFileUploadRequest) (StartFileUploadResponse, error) {
	if workspaceID == "" {
		return StartFileUploadResponse{}, fmt.Errorf("workspace ID is required")
	}
	if sessionID == "" {
		return StartFileUploadResponse{}, fmt.Errorf("session ID is required")
	}
	var resp StartFileUploadResponse
	p := wsPath(workspaceID, "/sessions/"+url.PathEscape(sessionID)+"/start-file-upload")
	if err := c.sendJSON(ctx, http.MethodPost, p, req, &resp); err != nil {
		return StartFileUploadResponse{}, err
	}
	return resp, nil
}

// SessionCompleteFileUpload finalizes an upload — the backend extracts the
// archive into the destination folder on the session VM.
// Endpoint: POST /v1/workspaces/{workspaceId}/sessions/{sessionId}/complete-file-upload.
func (c *Client) SessionCompleteFileUpload(ctx context.Context, workspaceID, sessionID string, req CompleteFileUploadRequest) error {
	if workspaceID == "" {
		return fmt.Errorf("workspace ID is required")
	}
	if sessionID == "" {
		return fmt.Errorf("session ID is required")
	}
	p := wsPath(workspaceID, "/sessions/"+url.PathEscape(sessionID)+"/complete-file-upload")
	return c.sendJSON(ctx, http.MethodPost, p, req, nil)
}

// SessionDownloadFile requests a signed GET URL for a tar.gz of remote files.
// The actual GET and untar are the caller's responsibility (internal/rde).
// Endpoint: POST /v1/workspaces/{workspaceId}/sessions/{sessionId}/download-file.
func (c *Client) SessionDownloadFile(ctx context.Context, workspaceID, sessionID string, req DownloadFileRequest) (DownloadFileResponse, error) {
	if workspaceID == "" {
		return DownloadFileResponse{}, fmt.Errorf("workspace ID is required")
	}
	if sessionID == "" {
		return DownloadFileResponse{}, fmt.Errorf("session ID is required")
	}
	if req.SourcePath == "" {
		return DownloadFileResponse{}, fmt.Errorf("source path is required")
	}
	var resp DownloadFileResponse
	p := wsPath(workspaceID, "/sessions/"+url.PathEscape(sessionID)+"/download-file")
	if err := c.sendJSON(ctx, http.MethodPost, p, req, &resp); err != nil {
		return DownloadFileResponse{}, err
	}
	return resp, nil
}
