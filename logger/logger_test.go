package logger

import "testing"

const path = `/mnt/d/project/go/src/log/file/log.log`

func TestNewStdLogger(t *testing.T) {
	l := NewStdLogger(path, true, true, true)

	l.Trace("%s, %s", "hello", "world!")
}
