package docker

import (
	"bytes"
	"fmt"
	"io"

	"github.com/bitrise-io/bitrise/v2/log"
	"github.com/bitrise-io/go-utils/v2/redactwriter"
)

type Logger struct {
	logger  log.Logger
	secrets []string
}

func (dl *Logger) Infof(format string, args ...interface{}) {
	str := fmt.Sprintf(format, args...)
	redacted, _ := dl.Redact(str)
	dl.logger.Info(redacted)
}

func (dl *Logger) Errorf(format string, args ...interface{}) {
	str := fmt.Sprintf(format, args...)
	redacted, _ := dl.Redact(str)
	dl.logger.Error(redacted)
}

func (dl *Logger) Warnf(format string, args ...interface{}) {
	str := fmt.Sprintf(format, args...)
	redacted, _ := dl.Redact(str)
	dl.logger.Warn(redacted)
}

func (dl *Logger) Redact(s string) (string, error) {
	src := bytes.NewReader([]byte(s))
	dstBuf := new(bytes.Buffer)
	logger := log.NewUtilsLogAdapter()
	redactWriterDst := redactwriter.New(dl.secrets, dstBuf, &logger)

	if _, err := io.Copy(redactWriterDst, src); err != nil {
		return "", fmt.Errorf("failed to redact secrets, stream copy failed: %s", err)
	}
	if err := redactWriterDst.Close(); err != nil {
		return "", fmt.Errorf("failed to redact secrets, closing the stream failed: %s", err)
	}

	redactedValue := dstBuf.String()
	return redactedValue, nil
}
