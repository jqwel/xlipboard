//go:build !no_fuse
// +build !no_fuse

package application

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/jqwel/xlipboard/src/utils/logger"

	"github.com/jqwel/xlipboard/src/rpc/iconst"

	"github.com/winfsp/cgofuse/fuse"

	"github.com/jqwel/xlipboard/src/utils"
)

var host *fuse.FileSystemHost

func StartMount(c *Config) {
	dir := filepath.ToSlash(filepath.Join(os.TempDir(), iconst.MountFolder))
	c.Mount = filepath.ToSlash(filepath.Join(dir, utils.RandStringBytes(8)))
	if err := os.RemoveAll(dir); err != nil {
		logger.Logger.Errorln(err)
	}
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		logger.Logger.Error(err)
		return
	}
	osType := runtime.GOOS
	switch osType {
	case "windows":
	case "darwin":
		err := os.MkdirAll(c.Mount, 0755)
		if err != nil {
			logger.Logger.Error(err)
			return
		}
	case "linux":
		err := os.MkdirAll(c.Mount, 0755)
		if err != nil {
			logger.Logger.Error(err)
			return
		}
	default:
		logger.Logger.Errorln("Unsupported operating system")
	}
	fmt.Println("Mounted at", c.Mount)
	gfs := &XlipboardFs{}
	host = fuse.NewFileSystemHost(gfs)
	host.Mount(c.Mount, nil)
}

func BeforeUnMount() {
	if host != nil {
		host.Unmount()
	}
	if App.Config != nil {
		if err := os.RemoveAll(App.Config.Mount); err != nil {
			logger.Logger.Errorln(err)
		}
	}
}
