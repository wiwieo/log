package normal

import (
	"fmt"
	"os"
	"time"
)

type normal struct {
	FilePath string
	f        *os.File
}

func New(filePath string) (*normal, error) {
	os.Rename(filePath, fmt.Sprintf("%s.%+v", filePath, time.Now().Format("20060102150405")))
	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	return &normal{
		FilePath: filePath,
		f:        f,
	}, nil
}

func (n *normal) Write(content []byte) error {
	if n.f == nil {
		return fmt.Errorf("file is not opened")
	}

	fi, err := n.f.Stat()
	if nil != err {
		return err
	}

	if _, err := n.f.WriteAt(content, fi.Size()); nil != err {
		return err
	}

	return nil
}

func (n *normal) Close() error {
	if n.f != nil {
		return n.f.Close()
	}
	return nil
}
