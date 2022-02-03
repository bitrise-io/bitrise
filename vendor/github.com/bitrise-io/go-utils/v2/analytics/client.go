package analytics

import (
	"bytes"
	"net/http"
	"time"

	"github.com/bitrise-io/go-utils/retry"
	"github.com/bitrise-io/go-utils/v2/log"
)

const trackEndpoint = "https://bitrise-step-analytics.herokuapp.com/track"
const timeOut = time.Second * 30

// Client ...
type Client interface {
	Send(buffer *bytes.Buffer)
}

type client struct {
	httpClient *http.Client
	endpoint   string
	logger     log.Logger
}

// NewDefaultClient ...
func NewDefaultClient(logger log.Logger) Client {
	httpClient := retry.NewHTTPClient().StandardClient()
	httpClient.Timeout = timeOut
	return NewClient(httpClient, trackEndpoint, logger)
}

// NewClient ...
func NewClient(httpClient *http.Client, endpoint string, logger log.Logger) Client {
	return client{httpClient: httpClient, endpoint: endpoint, logger: logger}
}

// Send ...
func (t client) Send(buffer *bytes.Buffer) {
	res, err := t.httpClient.Post(t.endpoint, "application/json", buffer)
	if err != nil {
		t.logger.Debugf("Couldn't send analytics event: %s", err.Error())
		return
	}
	if statusOK := res.StatusCode >= 200 && res.StatusCode < 300; !statusOK {
		t.logger.Debugf("Couldn't send analytics event, status code: %d", res.StatusCode)
	}
}
