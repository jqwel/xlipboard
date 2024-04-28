package utils

import (
	"sync"
	"time"

	"github.com/beevik/ntp"

	"github.com/jqwel/xlipboard/src/utils/logger"
)

var offset time.Duration
var fnMu sync.Mutex

func InitOffsetPeriodically(address string, interval time.Duration) error {
	if err := InitOffset(address); err != nil {
		return err
	}

	// 启动后台 goroutine 定时更新偏移量
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := InitOffset(address); err != nil {
					logger.Logger.Error(err)
				}
			}
		}
	}()
	return nil
}

var InitOffset = func(address string) error {
	response, err := ntp.QueryWithOptions(address, ntp.QueryOptions{Timeout: time.Second * 3})
	if err != nil {
		return err
	} else {
		fnMu.Lock()
		defer fnMu.Unlock()
		offset = response.ClockOffset
	}
	return nil
}

func GetFixedNow() time.Time {
	fnMu.Lock()
	defer fnMu.Unlock()
	return time.Now().Add(offset)
}
