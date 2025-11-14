package logger

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
