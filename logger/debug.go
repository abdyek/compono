//go:build debug

package logger

import (
	"fmt"
	"strings"
)

type logger struct {
	level       LogLevel
	indentLevel int
}

func NewLogger() Logger {
	return &logger{}
}

func (l *logger) SetLogLevel(level LogLevel) {
	l.level = level
}

func (l *logger) Log(level LogLevel, msg string, args ...interface{}) {
	if l.level&level != 0 {
		fmt.Printf("%s%s %s\n",
			strings.Repeat("  ", l.indentLevel),
			Bold(l.prefix(level)),
			fmt.Sprintf(msg, args...),
		)
	}
}

func (l *logger) LogMultiline(level LogLevel, msg string, args ...interface{}) {
	if l.level&level == 0 {
		return
	}

	formatted := fmt.Sprintf(msg, args...)
	lines := strings.Split(formatted, "\n")

	prefix := Bold(l.prefix(level))
	indent := strings.Repeat("  ", l.indentLevel)

	for i, line := range lines {
		if i == 0 {
			fmt.Printf("%s%s %s\n", indent, prefix, line)
		} else {
			padding := strings.Repeat(" ", len(prefix)+1)
			fmt.Printf("%s%s%s\n", indent, padding, line)
		}
	}
}

func (l *logger) Enter(level LogLevel, msg string, args ...interface{}) {
	l.Log(level, msg, args...)
	l.indentLevel++
}

func (l *logger) Exit(level LogLevel, msg string, args ...interface{}) {
	l.indentLevel--
	l.Log(level, msg, args...)
}

func (l *logger) prefix(level LogLevel) string {
	switch {
	case level&Parser != 0:
		if level&Detail != 0 {
			return "[PARSER:DETAIL]"
		}
		return "[PARSER]"
	case level&Renderer != 0:
		if level&Detail != 0 {
			return "[RENDERER:DETAIL]"
		}
		return "[RENDERER]"
	default:
		return "[LOG]"
	}
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
