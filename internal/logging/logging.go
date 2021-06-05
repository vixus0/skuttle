package logging

import (
	"fmt"
	"log"
	"os"
)

type LogLevel int

type Logger struct {
	*log.Logger
}

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

var (
	globalLevel LogLevel = INFO
)

func SetLevel(l LogLevel) {
	globalLevel = l
}

func GetLevel() LogLevel {
	return globalLevel
}

func (level LogLevel) String() string {
	switch level {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	}
	return "UNKNOWN"
}

func NewLogger(prefix string) *Logger {
	return &Logger{
		log.New(os.Stdout, fmt.Sprintf("%s: ", prefix), log.LstdFlags|log.Lmsgprefix|log.LUTC),
	}
}

func (l *Logger) Debug(format string, v ...interface{}) {
	l.Log(DEBUG, format, v...)
}

func (l *Logger) Info(format string, v ...interface{}) {
	l.Log(INFO, format, v...)
}

func (l *Logger) Warn(format string, v ...interface{}) {
	l.Log(WARN, format, v...)
}

func (l *Logger) Error(format string, v ...interface{}) {
	l.Log(ERROR, format, v...)
}

func (l *Logger) Log(level LogLevel, format string, v ...interface{}) {
	if level >= GetLevel() {
		levelFormat := fmt.Sprintf("[%s] %s", level, format)
		l.Printf(levelFormat, v...)
	}
}

func (l *Logger) Fatal(obj interface{}) {
	l.Fatalf("%v", obj)
}

func (l *Logger) Fatalf(format string, v ...interface{}) {
	levelFormat := fmt.Sprintf("[%s] %s", "FATAL", format)
	l.Printf(levelFormat, v...)
	os.Exit(1)
}
