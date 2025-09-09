package log

// HTTPLogAdaptor adapts the retryablehttp.Logger interface to the log.Logger.
type HTTPLogAdaptor struct {
	Logger Logger
}

// Printf implements the retryablehttp.Logger interface
func (a *HTTPLogAdaptor) Printf(fmtStr string, vars ...interface{}) {
	a.Logger.Debugf(fmtStr, vars...)
}
