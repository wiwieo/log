package main

import (
	"log/logger"
	"testing"
)

func TestNewLogger(t *testing.T) {
	l := logger.NewStdLogger(*logPath, true, true, true, true)
	for i := 0;i < 1000000; i ++{
		l.Trace("this is a trace log, %d", i)
	}
}
