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

func Colorize(text, color string) string {
	return ""
}

func Bold(text string) string {
	return ""
}

func Underline(text string) string {
	return ""
}

func BoldUnderline(text string) string {
	return ""
}

func Italic(text string) string {
	return ""
}

func Highlight(source []byte, indexes [][2]int, colorFunc func(string) string) string {
	return ""
}
