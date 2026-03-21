package log

import "strings"

type LogLevel uint8

const (
	RFU2    LogLevel = 255
	TRACE   LogLevel = 200
	DEBUG   LogLevel = 100
	INFO    LogLevel = 20
	WARN    LogLevel = 10
	ERROR   LogLevel = 5
	DISABLE LogLevel = 1
	RFU1    LogLevel = 0
)

type Logger struct {
	LogLevel LogLevel
	message  string
}

func ToLevel(l string) (LogLevel, bool) {
	l = strings.ToUpper(l)
	logLevel := map[string]LogLevel{
		"TRACE":   TRACE,
		"DEBUG":   DEBUG,
		"INFO":    INFO,
		"WARN":    WARN,
		"ERROR":   ERROR,
		"DISABLE": DISABLE,
	}

	level, ok := logLevel[l]
	if !ok {
		level = DEBUG
	}
	return level, ok
}

func kvPrint(kv ...any) string {
	var m strings.Builder
	for i := 0; i+1 < len(kv); i += 2 {
		m.WriteString(" " + kv[i].(string) + "=" + kv[i+1].(string))
	}

	print(m)
	return m.String()
}

// TODO
func (l *Logger) With(kv ...any) *Logger {
	l.message += kvPrint(kv...)
	return l
}

func (l *Logger) Debug(msg string, kv ...any) {
	if l.LogLevel >= DEBUG {
		var m strings.Builder

		m.WriteString("[DBG] " + msg)
		m.WriteString(kvPrint(kv...))
		m.WriteString("\r\n")

		print(m)
		l.message = m.String()
	}
}

func (l *Logger) Info(msg string, kv ...any) {
	if l.LogLevel >= INFO {
		var m strings.Builder

		m.WriteString("[INF] " + msg)
		m.WriteString(kvPrint(kv...))
		m.WriteString("\r\n")

		print(m)
		l.message = m.String()
	}
}

func (l *Logger) Warn(msg string, kv ...any) {
	if l.LogLevel >= WARN {
		var m strings.Builder

		m.WriteString("[WRN] " + msg)
		m.WriteString(kvPrint(kv...))
		m.WriteString("\r\n")

		print(m)
		l.message = m.String()
	}
}

func (l *Logger) Error(msg string, kv ...any) {
	if l.LogLevel >= ERROR {
		var m strings.Builder

		m.WriteString("[ERR] " + msg)
		m.WriteString(kvPrint(kv...))
		m.WriteString("\r\n")

		print(m)
		l.message = m.String()
	}
}
