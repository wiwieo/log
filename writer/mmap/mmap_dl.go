// +build linux,cgo darwin,cgo

package mmap

import (
	"fmt"
	"os"
	"syscall"
	"time"
)

func MmapRead(filePath string) (content []byte, err error) {
	// open a file
	fd, err := syscall.Open(filePath, syscall.O_RDONLY, 0)
	defer syscall.Close(fd)
	if err != nil {
		return nil, err
	}
	// get file size
	fsize, err := syscall.Seek(fd, 0, 2)
	if err != nil {
		return nil, err
	}
	content = make([]byte, fsize)
	content, err = syscall.Mmap(fd, 0, int(fsize), syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		return nil, err
	}
	return
}

type mmap struct {
	data       []byte      // 与文件映射的内存
	dataC      chan []byte // 用于写入的通道
	f          *os.File    // 日志文件
	FilePath   string      // 文件路径
	at         int         // 在什么位置写
	size       int         // 与文件映射的大小
	extendType int         // 当映射的内存不够用时的扩容策略，1: 在原文件中继续追加，2：新建文件
	isM        bool        // 如果使用mmap出错，则改用直接写文件的方式
}

func NewMmap(filePath string, size int, extendType int) (*mmap, error) {
	// 文件映射的大小必须是页数的倍数，如果不是，则自动根据大小调整为相应倍数
	if size%syscall.Getpagesize() != 0 {
		size = (size / syscall.Getpagesize()) * syscall.Getpagesize()
	}
	if size == 0 {
		size = syscall.Getpagesize()
	}

	// 构建对应的结构体，以配后续使用
	m := &mmap{
		size:       size,
		FilePath:   filePath,
		extendType: extendType,
		dataC:      make(chan []byte, 10),
		isM:        true,
	}

	// 使用channel方式，同步写入
	go m.wait()
	//return m, m.init(filePath)
	return nil,fmt.Errorf("")
}

func (m *mmap) init(filePath string) error {
	os.Rename(filePath, fmt.Sprintf("%s.%+v", filePath, time.Now().Format("20060102150405")))
	err := m.setFileInfo(filePath)
	if err != nil {
		return err
	}

	err = m.allocate()
	if err != nil {
		return err
	}
	return nil
}

func (m *mmap) allocate() error {
	if m.f == nil {
		m.setFileInfo(m.FilePath)
	}
	defer func() {
		m.f.Close()
		m.f = nil
	}()

	// MMAP映射时，文件必须有相应大小的内容，即需要相应大小的占位符
	if _, err := m.f.WriteAt(make([]byte, m.size), int64(m.at)); nil != err {
		return err
	}

	// 映射
	data, err := syscall.Mmap(int(m.f.Fd()), 0, int(m.size), syscall.PROT_WRITE|syscall.PROT_READ, syscall.MAP_SHARED)
	if nil != err {
		return err
	}
	m.data = data
	return nil
}

func (m *mmap) setFileInfo(filePath string) error {
	// 打开文件
	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0644)
	if nil != err {
		return err
	}

	// 获取当前文件信息
	fi, err := f.Stat()
	if nil != err {
		return err
	}

	m.f = f
	m.at = int(fi.Size())
	return nil
}

// 关闭文件映射
func (m *mmap) Close() error {
	// 关闭映射
	if err := syscall.Munmap(m.data); nil != err {
		return err
	}

	// 将未写入的内容清空
	// 如果未清空，在文件末位未写入位置，将会出现大量占位符
	err := os.Truncate(m.FilePath, int64(m.at))
	if err != nil {
		return err
	}
	return nil
}

// 接收写入内容
func (m *mmap) Write(content []byte) error {
	m.dataC <- content
	return nil
}

// 当初始映射大小不足时，以双倍方式扩容
func (m *mmap) doubleMmap() error {
	// 先将之前的映射关闭
	m.Close()
	m.size = 2 * m.size
	return m.allocate()
}

func (m *mmap) wait() {
	for {
		select {
		case content := <-m.dataC:
			if len(content) == 0 {
				return
			}
			// 剩余空间不足以添加所有内容，需要扩容
			for len(content) > m.size-m.at {
				err := m.doubleMmap()
				if err != nil {
					m.isM = false
				}
			}
			m.write(content)
		}
	}
}

func (m *mmap) write(content []byte) {
	if m.isM {
		m.writeWithMmap(content)
	} else {
		m.writeWithIO(content)
	}
}

func (m *mmap) writeWithMmap(content []byte) {
	// 内容写入文件
	for i, v := range content {
		m.data[m.at+i] = v
	}
	m.at += len(content)
}

func (m *mmap) writeWithIO(content []byte) {
	var err error
	if m.f == nil {
		m.f, err = os.OpenFile(m.FilePath, os.O_RDWR|os.O_CREATE, os.ModeAppend)
		if err != nil {
			panic(err)
		}
	}

	size, err := m.f.WriteAt(content, int64(m.at))
	m.at += size
	if err != nil {
		panic(err)
	}
}
