package logger

type LogLevel uint8

const (
	None   LogLevel = 0
	Parser LogLevel = 1 << iota
	Renderer
	Detail
)

const All = Parser | Renderer | Detail

type Logger interface {
	SetLogLevel(level LogLevel)
	Log(level LogLevel, msg string, args ...interface{})
	Enter(level LogLevel, msg string, args ...interface{})
	Exit(level LogLevel, msg string, args ...interface{})
}
