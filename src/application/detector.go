package application

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/jqwel/xlipboard/src/rpc/iconst"
	"github.com/jqwel/xlipboard/src/utils/logger"

	"github.com/jqwel/xlipboard/src/rpc/xlipboard/client"
	"github.com/jqwel/xlipboard/src/utils"
)

type Detector struct {
}

func NewDetector() *Detector {
	return &Detector{}
}

func (d *Detector) Start() {
	go d.StartFetch()
}

type SeqHa struct {
	Seq     string
	Ha      string
	Init    bool
	Virtual bool
}

var SH = SeqHa{}

func (d *Detector) check() error {
	contentType, copyStr, copyImageByte, copyFilename, err := getContentTypeAndData()
	if err != nil {
		//return err
		contentType = utils.TypeText // empty
	}

	seq, err := utils.Xlipboard().ClipboardSequence()
	if SH.Seq == seq && utils.InSlice(contentType, []string{utils.TypeText, utils.TypeBitmap}) {
		return nil
	}
	SH.Seq = seq

	newHa := contentType + "|"

	var virtual = false
	if contentType == utils.TypeText {
		hash := md5.Sum([]byte(copyStr))
		hashString := hex.EncodeToString(hash[:])
		newHa += hashString
	} else if contentType == utils.TypeBitmap {
		hash := md5.Sum([]byte(copyImageByte))
		hashString := hex.EncodeToString(hash[:])
		newHa += hashString
	} else if contentType == utils.TypeFile {
		sort.Strings(copyFilename)
		checkInput := strings.Join(copyFilename, ";")
		for _, filePath := range copyFilename {
			if runtime.GOOS == "windows" {
				if strings.Index(filepath.ToSlash(os.TempDir()), "AppData/Local/Temp") > 0 && strings.Index(filepath.ToSlash(filePath), "AppData/Local/Temp") > 0 {
					virtual = true
				}
			} else {
				if !virtual && strings.Index(filepath.ToSlash(filePath), filepath.ToSlash(App.Config.Mount)) == 0 {
					virtual = true
				}
			}
			info, err := os.Stat(filePath)
			if err != nil {
				continue
			}
			ts := fmt.Sprintf("|%s|%d|%d|", info.Name(), info.Size(), info.ModTime().UnixNano())
			checkInput += ts
		}
		hash := md5.Sum([]byte(checkInput))
		hashString := hex.EncodeToString(hash[:])
		newHa += hashString
	}

	if !SH.Init {
		SH.Init = true
		return nil
	}
	SH.Virtual = virtual
	if SH.Ha == newHa {
		return nil
	}
	SH.Ha = newHa
	now := utils.GetFixedNow()
	changeAt := now.UnixMilli()
	if changeAt > App.GetChangeAt() {
		App.SetChangeAt(changeAt)
		App.SetVirtual(SH.Virtual)
	}
	return nil
}

func (d *Detector) StartFetch() {
	for {
		if err := d.fetch(); err != nil {
			logger.Logger.Errorln(err)
		}
		time.Sleep(time.Millisecond * 600)
	}
}

var resetMu sync.Mutex

func ResetSH(changeAt int64, virtual bool) {
	resetMu.Lock()
	defer resetMu.Unlock()
	SH = SeqHa{
		Virtual: virtual,
	}
	App.SetChangeAt(changeAt)
	App.SetVirtual(virtual)
}

type HelloResult struct {
	Target   string
	ChangeAt int64
	Virtual  bool
	Now      int64
}

func (d *Detector) fetch() error {
	go d.check()

	if len(App.Config.SyncSettings) == 0 {
		return nil
	}
	resetMu.Lock()
	defer resetMu.Unlock()

	changeAt := App.GetChangeAt()
	client.InitSettings(App.Config.Authkey, App.Config.Certificate, App.Config.PrivateKey)

	var wg sync.WaitGroup
	resultChan := make(chan *HelloResult, len(App.Config.SyncSettings))

	for _, setting := range App.Config.SyncSettings {
		target := setting.Target
		if target == "" {
			continue
		}
		wg.Add(1)
		go func(target string, resultChan chan<- *HelloResult, wg *sync.WaitGroup) {
			defer wg.Done()
			// 获取目标服务器的changeAt
			shr, err := client.SayHello(target, changeAt)
			if err != nil {
				if err.Error() == iconst.Timeout {
					logger.Logger.Warning("SayHello", err, target)
				}
				return
			}
			resultChan <- &HelloResult{Target: target, ChangeAt: shr.GetTimestamp(), Virtual: shr.GetVirtual(), Now: shr.GetNow()}
		}(target, resultChan, &wg)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	var helloResult = HelloResult{
		Target:   "",
		ChangeAt: App.GetChangeAt(),
	}
	for result := range resultChan {
		if result.ChangeAt > App.GetChangeAt() && !result.Virtual && result.Target != "" {
			helloResult = *result
		}
	}
	if helloResult.Target == "" {
		return nil
	}
	if math.Abs(float64(helloResult.Now-utils.GetFixedNow().UnixMilli())) > 2000 {
		go utils.InitOffsetPeriodically(App.Config.NtpAddress, 0)
	}

	shrur, err := client.SayHowAreYou(helloResult.Target, helloResult.ChangeAt)
	if err != nil {
		return err
	}
	if helloResult.ChangeAt != shrur.GetTimestamp() {
		helloResult.ChangeAt = shrur.GetTimestamp()
	}
	switch shrur.GetContentType() {
	case utils.TypeText:
		if App.IsDebug {
			logger.Logger.Debugln("文本内容是:", shrur.GetCopyStr())
		}
		if err := utils.Xlipboard().SetText(shrur.GetCopyStr()); err != nil {
			return err
		}
		go ResetSH(helloResult.ChangeAt, false)
		return nil
	case utils.TypeBitmap:
		return errors.New("invalid type ...")
	case utils.TypeFile:
		d.SetFilename(helloResult.Target, shrur.GetTimestamp(), shrur.GetCopyFilename())
		go ResetSH(helloResult.ChangeAt, true)
		return nil
	default:
		go ResetSH(helloResult.ChangeAt, true)
		return errors.New("wrong content type")
	}
}

func (d *Detector) SetFilename(target string, timestamp int64, filenames []string) error {
	fns, err := AddFilenames(target, timestamp, filenames)
	if err != nil {
		return err
	}
	if fns != nil {
		return utils.Xlipboard().SetFiles(fns)
	}
	return nil
}
