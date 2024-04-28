package application

import (
	"sync"
	"sync/atomic"

	"github.com/jqwel/xlipboard/src/utils/logger"
)

type CacheBuffer struct {
	start uint64
	end   uint64
	data  []byte
	count uint32
	ch    chan struct{}
}

type CacheBufferMap = map[uint64]*CacheBuffer // key: start

type XlipboardCache struct {
	mu    sync.Mutex
	cache map[uint64]CacheBufferMap // fh
}

func (qc *XlipboardCache) InitCa(fh uint64) {
	qc.mu.Lock()
	defer qc.mu.Unlock()
	if qc.cache == nil {
		qc.cache = make(map[uint64]CacheBufferMap)
	}
	qc.cache[fh] = make(CacheBufferMap)
}
func (qc *XlipboardCache) ClearCa(fh uint64) {
	qc.mu.Lock()
	defer qc.mu.Unlock()
	if bufferMap, ok := qc.cache[fh]; ok {
		if bufferMap != nil {
			go func() {
				for _, buffer := range bufferMap {
					if buffer.ch != nil {
						close(buffer.ch)
					}
				}
			}()
		}
	}
	delete(qc.cache, fh)
}

func (qc *XlipboardCache) CheckCaRequested(fh uint64, ofstFix int64, lenbuff int, mul uint32) bool {
	var count uint32 = 1
	ofst := ofstFix
	if ofst < 0 {
		count = 0
		ofst = -ofst
	}
	batchSize := int64(lenbuff) * int64(mul)
	n := ofst / batchSize
	keyStart := uint64(n * batchSize)
	keyEnd := uint64(n*batchSize + batchSize)

	qc.mu.Lock()
	defer qc.mu.Unlock()
	result := true
	if _, ok := qc.cache[fh]; !ok {
		qc.cache[fh] = make(CacheBufferMap)
	}
	buffer, ok := qc.cache[fh][keyStart]
	if !ok || buffer == nil {
		qc.cache[fh][keyStart] = &CacheBuffer{
			start: keyStart,
			end:   keyEnd,
			data:  nil,
			ch:    make(chan struct{}, 1),
			count: count,
		}
		result = false
	}
	return result
}

func (qc *XlipboardCache) ResultCaRequested(fh uint64, ofst int64, lenbuff int, mul uint32, read int32, dataMul []byte) {
	qc.mu.Lock()
	defer qc.mu.Unlock()
	batchSize := int64(lenbuff) * int64(mul)
	n := ofst / batchSize
	keyStart := uint64(n * batchSize)
	if bufferMap, ok := qc.cache[fh]; !ok || bufferMap == nil {
		// cleared
		return
	}

	cb := qc.cache[fh][keyStart]
	if mul == cb.count {
		return
	}
	cb.data = dataMul
	cb.ch <- struct{}{}
}

func (qc *XlipboardCache) ResultCaFetch(fh uint64, ofstFix int64, lenbuff int, mul uint32) []byte {
	ofst := ofstFix
	if ofst < 0 {
		ofst = -ofst
	}
	batchSize := int64(lenbuff) * int64(mul)
	n := ofst / batchSize
	keyStart := uint64(n * batchSize)
	qc.mu.Lock()
	if bufferMap, ok := qc.cache[fh]; !ok || bufferMap == nil {
		qc.mu.Unlock()
		return nil
	}
	cb := qc.cache[fh][keyStart]
	qc.mu.Unlock()
	<-cb.ch
	defer func() {
		cb.ch <- struct{}{}
	}()
	if cb.data == nil {
		return nil
	}
	chunk := cb.data
	chunkStart := ofst - int64(keyStart)
	chunkEnd := chunkStart + int64(lenbuff)
	if chunkEnd > int64(len(chunk)) {
		chunkEnd = int64(len(chunk))
	}
	if ofstFix >= 0 {
		atomic.AddUint32(&cb.count, 1)
	} else {
		return nil // 优化返回
	}
	qc.mu.Lock()
	defer qc.mu.Unlock()
	if cb.count == mul {
		cb.data = nil
	}
	if cb.count > mul {
		logger.Logger.Error("ResultCaFetch: cb.data = nil")
	}
	return chunk[chunkStart:chunkEnd]
}
