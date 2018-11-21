package main

import (
	"flag"
	_ "github.com/mkevac/debugcharts"
	"log/logger"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	logPath = flag.String("log_path", `/mnt/d/project/go/src/log/file/log.log`, "absolute file path")
)

func init() {
	flag.Parse()
}

type server struct {
	logger *logger.Logger
}

func main() {
	l := logger.NewStdLogger(*logPath, true, true, true)
	s := &server{
		logger: l,
	}
	go func() {
		t := time.NewTicker(100 * time.Millisecond)
		for {
			select {
			case <-t.C:
				s.logger.TraceWithField(map[string]interface{}{"hello": "你好。", "hi": "こんにちは！"}, "now:[%+v] this is a trace log", time.Now().Format("2006-01-02 15:04:05.9999"))
			}
		}
	}()
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Kill, os.Interrupt, syscall.SIGUSR1, syscall.SIGUSR2)
	go func() {
		<-ch
		l.Close()
		os.Exit(1)
	}()
	http.HandleFunc("/hello", s.Hello)
	http.ListenAndServe(":8888", nil)
}

func (s *server) Hello(w http.ResponseWriter, r *http.Request) {
	for i := 0; i < 100000; i++ {
		s.logger.Trace("this is a trace log")
	}
	w.Write([]byte("success"))
}
