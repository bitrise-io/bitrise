package analytics

import (
	"bytes"
	"context"
	"net/http"
	"time"

	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-utils/v2/retryhttp"
)

const trackEndpoint = "https://step-analytics.bitrise.io/track"

// Client ...
type Client interface {
	Send(buffer *bytes.Buffer)
}

type client struct {
	httpClient *http.Client
	timeout    time.Duration
	endpoint   string
	logger     log.Logger
}

// NewDefaultClient ...
func NewDefaultClient(logger log.Logger, timeout time.Duration) Client {
	httpClient := retryhttp.NewClient(logger).StandardClient()
	httpClient.Timeout = timeout
	return NewClient(httpClient, trackEndpoint, logger, timeout)
}

// NewClient ...
func NewClient(httpClient *http.Client, endpoint string, logger log.Logger, timeout time.Duration) Client {
	return client{httpClient: httpClient, endpoint: endpoint, logger: logger, timeout: timeout}
}

// Send ...
func (t client) Send(buffer *bytes.Buffer) {
	ctx, cancel := context.WithTimeout(context.Background(), t.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.endpoint, buffer)
	if err != nil {
		t.logger.Warnf("Couldn't create analytics request: %s", err)
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := t.httpClient.Do(req)
	if err != nil {
		t.logger.Debugf("Couldn't send analytics event: %s", err)
		return
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			t.logger.Debugf("Couldn't close anaytics body: %s", err)
		}
	}()

	if statusOK := res.StatusCode >= 200 && res.StatusCode < 300; !statusOK {
		t.logger.Debugf("Couldn't send analytics event, status code: %d", res.StatusCode)
	}
}
