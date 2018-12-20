package log

import (
	"testing"
	"uuabc.com/gateway/pkg/log/logger"
)

func TestNewLogger(t *testing.T) {
	l := logger.NewStdLogger(true, true, true, true, true)
	l.SetPath(*logPath)
	for i := 0; i < 1000000; i++ {
		l.Trace("this is a trace log, %d", i)
	}
}
