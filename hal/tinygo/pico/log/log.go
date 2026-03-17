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

func kvPrint(kv ...any) {
	for i := 0; i+1 < len(kv); i += 2 {
		if i%2 != 0 {
			break
		}

		print(" ")
		print(kv[i])
		print("=")
		print(kv[i+1])
	}
}

func (l *Logger) With(kv ...any) *Logger { return l } // TODO

func (l *Logger) Debug(msg string, kv ...any) {
	if l.LogLevel >= DEBUG {
		print("[DBG] " + msg)
		kvPrint(kv...)
		print("\r\n")
	}
}

func (l *Logger) Info(msg string, kv ...any) {
	if l.LogLevel >= INFO {
		print("[INF] " + msg)
		kvPrint(kv...)
		print("\r\n")
	}
}

func (l *Logger) Warn(msg string, kv ...any) {
	if l.LogLevel >= WARN {
		print("[WRN] " + msg)
		kvPrint(kv...)
		print("\r\n")
	}
}

func (l *Logger) Error(msg string, kv ...any) {
	if l.LogLevel >= ERROR {
		print("[ERR] " + msg)
		kvPrint(kv...)
		print("\r\n")
	}
}
