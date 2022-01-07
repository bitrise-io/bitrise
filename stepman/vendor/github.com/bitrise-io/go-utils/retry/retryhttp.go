package retry

import (
	"github.com/bitrise-io/go-utils/log"
	retryablehttp "github.com/hashicorp/go-retryablehttp"
)

// HTTPLogAdaptor adapts the retryablehttp.Logger interface to the go-utils logger.
type HTTPLogAdaptor struct{}

// Printf implements the retryablehttp.Logger interface
func (*HTTPLogAdaptor) Printf(fmtStr string, vars ...interface{}) {
	log.Debugf(fmtStr, vars...)
}

// NewHTTPClient returns a retryable HTTP client
func NewHTTPClient() *retryablehttp.Client {
	client := retryablehttp.NewClient()
	client.Logger = &HTTPLogAdaptor{}
	client.ErrorHandler = retryablehttp.PassthroughErrorHandler

	return client
}
