package server

import (
	"os"
	"sync"

	"github.com/jqwel/xlipboard/src/rpc/iconst"
	"github.com/jqwel/xlipboard/src/utils/logger"
)

const InitLenChanFile = 0  // 初始化长度 // 最大长度
const IncrementLenChan = 1 // 每次增加的长度

type FileMap struct {
	name    string
	mu      sync.Mutex
	ch      chan *os.File
	current int
	maximum int
	closed  bool
}

func NewFileMap(name string) *FileMap {
	ch := make(chan *os.File, iconst.MaxLenChanFile)
	for i := 0; i < InitLenChanFile; i++ {
		file, err := os.Open(name)
		if err != nil {
			logger.Logger.Error(err)
		}
		ch <- file
	}
	return &FileMap{
		name:    name,
		ch:      ch,
		current: InitLenChanFile,
		maximum: iconst.MaxLenChanFile,
	}
}

func (sm *FileMap) GetOne() (*os.File, bool, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	select {
	case file := <-sm.ch:
		return file, true, nil
	default:
		if sm.current < sm.maximum {
			sm.current += IncrementLenChan
			for i := 0; i < IncrementLenChan; i++ {
				file, err := os.Open(sm.name)
				if err != nil {
					logger.Logger.Error(err)
					return nil, false, err
				}
				sm.ch <- file
			}
			return <-sm.ch, true, nil
		}
	}
	return nil, false, nil
}

func (sm *FileMap) PutBack(file *os.File) {
	sm.ch <- file
}

func (sm *FileMap) Clear() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	for i := 0; i < sm.current; i++ {
		file := <-sm.ch
		file.Close()
	}
	sm.current = 0
}

func (sm *FileMap) MemClear() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	for i := 0; i < sm.current; i++ {
		file := <-sm.ch
		file.Close()
	}
	sm.current = 0
	if sm.closed {
		return
	}
	sm.closed = true
	close(sm.ch)
}
