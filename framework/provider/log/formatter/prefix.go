package formatter

import "github.com/hiholder/geex/framework/contract"

var prefixMap = map[contract.LogLevel]string {
	contract.FatalLevel: "[Fatal]",
	contract.WarnLevel: "[Warn]",
	contract.ErrorLevel: "[Error]",
	contract.InfoLevel: "[Info]",
	contract.DebugLevel: "[Debug]",
	contract.TraceLevel: "Trace",
}
func Prefix(level contract.LogLevel) string {
	return prefixMap[level]
}
