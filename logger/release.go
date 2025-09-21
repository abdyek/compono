//go:build !debug

package logger

type logger struct {
}

func NewLogger() Logger {
	return &logger{}
}

func (*logger) SetLogLevel(LogLevel)                          {}
func (*logger) Log(LogLevel, string, ...interface{})          {}
func (*logger) LogMultiline(LogLevel, string, ...interface{}) {}
func (*logger) Enter(LogLevel, string, ...interface{})        {}
func (*logger) Exit(LogLevel, string, ...interface{})         {}
