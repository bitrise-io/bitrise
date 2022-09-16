package retryhttp

import (
	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/hashicorp/go-retryablehttp"
)

// NewClient returns a retryable HTTP client with common defaults
func NewClient(logger log.Logger) *retryablehttp.Client {
	client := retryablehttp.NewClient()
	client.Logger = &httpLogAdaptor{logger: logger}
	client.ErrorHandler = retryablehttp.PassthroughErrorHandler

	return client
}

// httpLogAdaptor adapts the retryablehttp.Logger interface to the go-utils logger.
type httpLogAdaptor struct {
	logger log.Logger
}

// Printf implements the retryablehttp.Logger interface
func (a *httpLogAdaptor) Printf(fmtStr string, vars ...interface{}) {
	a.logger.Debugf(fmtStr, vars...)
}
