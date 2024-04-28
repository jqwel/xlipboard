package client

import (
	"math/rand"
	"sync"

	"google.golang.org/grpc"

	"github.com/jqwel/xlipboard/src/rpc/iconst"
	"github.com/jqwel/xlipboard/src/utils/logger"
)

const InitLenChanConn = 0  // 初始化长度 // 最大长度
const IncrementLenChan = 1 // 每次增加的长度

type ConnMap struct {
	name    string
	fn      func(target string) (*grpc.ClientConn, error)
	mu      sync.Mutex
	ch      chan *grpc.ClientConn
	current int
	maximum int
}

func NewConnMap(name string, fn func(target string) (*grpc.ClientConn, error)) *ConnMap {
	ch := make(chan *grpc.ClientConn, iconst.MaxLenChanConn)
	for i := 0; i < InitLenChanConn; i++ {
		file, err := fn(name)
		if err != nil {
			logger.Logger.Error(err)
		}
		ch <- file
	}
	return &ConnMap{
		name:    name,
		fn:      fn,
		ch:      ch,
		current: InitLenChanConn,
		maximum: iconst.MaxLenChanConn,
	}
}

func (sm *ConnMap) GetOne() (*grpc.ClientConn, bool, error) {
	select {
	case conn := <-sm.ch:
		return conn, true, nil
	default:
		sm.mu.Lock()
		defer sm.mu.Unlock()
		if sm.current < sm.maximum {
			sm.current += IncrementLenChan
			for i := 0; i < IncrementLenChan; i++ {
				file, err := sm.fn(sm.name)
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

func (sm *ConnMap) PutBack(conn *grpc.ClientConn, e error) {
	if e != nil || rand.Float64() > 0.99 {
		sm.mu.Lock()
		defer sm.mu.Unlock()
		sm.current = sm.current - 1
		conn.Close()
		return
	} else {
		sm.ch <- conn
	}
}
