// Copyright 2012-2015 Apcera Inc. All rights reserved.

//Package logger provides logging facilities for the NATS server
package logger

import (
	"fmt"
	"log/writer"
	"os"
)

// Logger is the server logger
type Logger struct {
	logger     writer.Writer
	debug      bool
	trace      bool
	warnLabel  string
	errorLabel string
	fatalLabel string
	debugLabel string
	traceLabel string
	pid        string
}

// NewStdLogger creates a logger with output directed to Stderr
func NewStdLogger(path string, debug, trace, colors, pid bool) *Logger {
	pre := ""
	if pid {
		pre = pidPrefix()
	}

	l := &Logger{
		logger:writer.NewWriter(path, 1 << 20, 0),
		debug: debug,
		trace: trace,
		pid:   pre,
	}

	if colors {
		setColoredLabelFormats(l)
	} else {
		setPlainLabelFormats(l)
	}

	return l
}

// Generate the pid prefix string
func pidPrefix() string {
	return fmt.Sprintf("[%d] ", os.Getpid())
}

func setPlainLabelFormats(l *Logger) {
	l.debugLabel = "[DBG] "
	l.traceLabel = "[TRC] "
	l.warnLabel = "[WAR] "
	l.errorLabel = "[ERR] "
	l.fatalLabel = "[FTL] "
}

func setColoredLabelFormats(l *Logger) {
	colorFormat := "[\x1b[%dm%s\x1b[0m] "
	l.debugLabel = fmt.Sprintf(colorFormat, 36, "DBG")
	l.traceLabel = fmt.Sprintf(colorFormat, 33, "TRC")
	l.warnLabel = fmt.Sprintf(colorFormat, 32, "WAR")
	l.errorLabel = fmt.Sprintf(colorFormat, 31, "ERR")
	l.fatalLabel = fmt.Sprintf(colorFormat, 31, "FTL")
}

// Debug logs a debug statement
func (l *Logger) Debug(format string, v ...interface{}) {
	if l.debug {
		l.logger.Write([]byte(fmt.Sprintf("%s %s %s", l.pid, l.debugLabel, fmt.Sprintf(format, v...))))
		l.logger.Write([]byte(fmt.Sprintln()))
	}
}

// Trace logs a trace statement
func (l *Logger) Trace(format string, v ...interface{}) {
	if l.trace {
		l.logger.Write([]byte(fmt.Sprintf("%s %s %s", l.pid, l.traceLabel, fmt.Sprintf(format, v...))))
		l.logger.Write([]byte(fmt.Sprintln()))
	}
}

// Warning logs a notice statement
func (l *Logger) Warning(format string, v ...interface{}) {
	l.logger.Write([]byte(fmt.Sprintf("%s %s %s", l.pid, l.warnLabel, fmt.Sprintf(format, v...))))
	l.logger.Write([]byte(fmt.Sprintln()))
}

// Error logs an error statement
func (l *Logger) Error(format string, v ...interface{}) {
	l.logger.Write([]byte(fmt.Sprintf("%s %s %s", l.pid, l.errorLabel, fmt.Sprintf(format, v...))))
	l.logger.Write([]byte(fmt.Sprintln()))
}

// Fatal logs a fatal error
func (l *Logger) Fatal(format string, v ...interface{}) {
	l.logger.Write([]byte(fmt.Sprintf("%s %s %s", l.pid, l.fatalLabel, fmt.Sprintf(format, v...))))
	l.logger.Write([]byte(fmt.Sprintln()))
}
