package application

import (
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/jqwel/xlipboard/src/utils"
)

type XlipFileRemote struct {
	Target          string
	Timestamp       int64
	Filenames       []string
	TimestampAccess int64
	mu              sync.Mutex
}

func (g *XlipFileRemote) FsAccessed() {
	go func() {
		g.mu.Lock()
		defer g.mu.Unlock()
		g.TimestampAccess = utils.GetFixedNow().UnixMilli()
	}()
}
func (g *XlipFileRemote) GetTimestampAccessed() int64 {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.TimestampAccess
}

const (
	MaxXlipFileRemotes = 256
)

var xlipFileRemotes = []*XlipFileRemote{}

var muXlipFileRemotes sync.Mutex

func AddFilenames(target string, timestamp int64, filenames []string) ([]string, error) {
	muXlipFileRemotes.Lock()
	muXlipFileRemotes.Unlock()
	xlipFileRemotes = append(xlipFileRemotes, &XlipFileRemote{
		Target:    target,
		Timestamp: timestamp,
		Filenames: filenames,
	})
	if len(xlipFileRemotes) > MaxXlipFileRemotes {
		xlipFileRemotes = xlipFileRemotes[len(xlipFileRemotes)-MaxXlipFileRemotes:]
	}
	var resultFilenames []string
	for i := range filenames {
		newFilename := fmt.Sprintf("%s/%d/%s", App.Config.Mount, timestamp, filepath.Base(filenames[i]))
		resultFilenames = append(resultFilenames, newFilename)
	}
	go GetFileList() // 优化
	return resultFilenames, nil
}
func GetFileList() []*XlipFileRemote {
	muXlipFileRemotes.Lock()
	muXlipFileRemotes.Unlock()
	var result []*XlipFileRemote
	for i := 0; i < len(xlipFileRemotes); i++ {
		p := xlipFileRemotes[i]
		if i >= len(xlipFileRemotes)-1 || p.GetTimestampAccessed() >= utils.GetFixedNow().Add(-1*time.Minute*5).UnixMilli() {
			result = append(result, p)
		}
	}
	xlipFileRemotes = result[:]
	return result
}
