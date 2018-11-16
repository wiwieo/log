package main

import (
	"flag"
	"log/logger"
	"time"
)

var (
	logPath = flag.String("log_path", `/mnt/d/project/go/src/log/file/log.log`, "absolute file path")
)

func init() {
	flag.Parse()
}

func main() {
	l := logger.NewStdLogger(*logPath, true, true, true, true)
	for {
		go func() {
			l.Trace("this is a trace log")
			time.Sleep(time.Second)
		}()
		go func() {
			l.Debug("this is a trace log")
			time.Sleep(time.Second)
		}()
		go func() {
			l.Error("this is a trace log")
			time.Sleep(time.Second)
		}()
		go func() {
			l.Warning("this is a trace log")
			time.Sleep(time.Second)
		}()
	}
}