package writer

import (
	"log/writer/mmap"
	"log/writer/normal"
)

type Writer interface {
	Write(content []byte) error
	Close() error
}

func NewWriter(filePath string, size int, extendType int) Writer {
	m, err := mmap.NewMmap(filePath, size, extendType)
	if err != nil {
		n, err := normal.New(filePath)
		if err != nil {
			panic(err)
		}
		return n
	}
	return m
}
