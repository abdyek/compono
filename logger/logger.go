package logger

import "fmt"

type LogLevel uint8

const (
	None   LogLevel = 0
	Parser LogLevel = 1 << iota
	Renderer
	Detail
)

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Cyan   = "\033[36m"
)

const All = Parser | Renderer | Detail

type Logger interface {
	SetLogLevel(level LogLevel)
	Log(level LogLevel, msg string, args ...interface{})
	LogMultiline(level LogLevel, msg string, args ...interface{})
	Enter(level LogLevel, msg string, args ...interface{})
	Exit(level LogLevel, msg string, args ...interface{})
}

func Colorize(text, color string) string {
	return color + text + Reset
}

func Bold(text string) string {
	return fmt.Sprintf("\033[1m%s\033[0m", text)
}

func Underline(text string) string {
	return fmt.Sprintf("\033[4m%s\033[0m", text)
}

func BoldUnderline(text string) string {
	return fmt.Sprintf("\033[1;4m%s\033[0m", text)
}

func Italic(text string) string {
	return fmt.Sprintf("\033[3m%s\033[0m", text)
}

func Highlight(source []byte, indexes [][2]int, colorFunc func(string) string) string {
	out := ""
	last := 0
	for _, idx := range indexes {
		out += string(source[last:idx[0]])
		out += colorFunc(string(source[idx[0]:idx[1]]))
		last = idx[1]
	}
	out += string(source[last:])
	return out
}
