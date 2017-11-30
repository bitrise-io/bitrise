package log

import "github.com/bitrise-io/go-utils/colorstring"

// Severity ...
type Severity uint8

const (
	errorSeverity Severity = iota
	warnSeverity
	normalSeverity
	infoSeverity
	successSeverity
	debugSeverity
)

type severityColorFunc colorstring.ColorfFunc

var (
	successSeverityColorFunc severityColorFunc = colorstring.Greenf
	infoSeverityColorFunc    severityColorFunc = colorstring.Bluef
	normalSeverityColorFunc  severityColorFunc = colorstring.NoColorf
	debugSeverityColorFunc   severityColorFunc = colorstring.NoColorf
	warnSeverityColorFunc    severityColorFunc = colorstring.Yellowf
	errorSeverityColorFunc   severityColorFunc = colorstring.Redf
)

var severityColorFuncMap = map[Severity]severityColorFunc{
	successSeverity: successSeverityColorFunc,
	infoSeverity:    infoSeverityColorFunc,
	normalSeverity:  normalSeverityColorFunc,
	debugSeverity:   debugSeverityColorFunc,
	warnSeverity:    warnSeverityColorFunc,
	errorSeverity:   errorSeverityColorFunc,
}
