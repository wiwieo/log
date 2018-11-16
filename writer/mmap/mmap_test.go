package mmap

import (
	"fmt"
	"os"
	"testing"
	"time"
)

const path  = `/mnt/d/project/go/src/log/file/log.log`

func TestMmapWrite(t *testing.T)  {
	m, err := NewMmap(path, 1 << 14, 1)
	if err != nil{
		panic(fmt.Sprintf("memory mapping to file error. %s", err))
	}
	for i:=0;i<1000000;i++ {
		err = m.Write([]byte(fmt.Sprintf("I haven't seen you for ages, %d.\r\n", i)))
		if err != nil {
			println(fmt.Sprintf("write to file failed. %s", err))
		}
	}
	time.Sleep(1*time.Second)
	m.Close()
}

func BenchmarkMmapWrite(b *testing.B) {
	m, err := NewMmap(path, 1 << 14, 1)
	if err != nil{
		panic(fmt.Sprintf("memory mapping to file error. %s", err))
	}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next(){
			err = m.Write([]byte("I haven't seen you for ages.\r\n"))
			if err != nil{
				println(fmt.Sprintf("write to file failed. %s", err))
				b.Fail()
			}
		}
	})
	m.Close()
}

func TestTruncate(t *testing.T)  {
	err := os.Truncate(path, 0)
	if err != nil{
		panic(err)
	}
}