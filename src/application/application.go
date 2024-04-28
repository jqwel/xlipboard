package application

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/jqwel/xlipboard/src/utils/logger"

	"github.com/energye/systray"
	"google.golang.org/grpc"

	"github.com/jqwel/xlipboard/src/rpc/iconst"
	"github.com/jqwel/xlipboard/src/static"
	"github.com/jqwel/xlipboard/src/utils"
)

type Application struct {
	Config   *Config
	ExecPath string
	Server   *grpc.Server
	sync.Mutex
	SyncStart bool
	ChangeAt  int64
	Virtual   bool
	IsDebug   bool
	mu        sync.Mutex
}

func (a *Application) SetChangeAt(changeAt int64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.ChangeAt = changeAt
}
func (a *Application) GetChangeAt() int64 {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.ChangeAt
}
func (a *Application) SetVirtual(virtual bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Virtual = virtual
}
func (a *Application) GetVirtual() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.Virtual
}

func (app *Application) GetAuthKey() string {
	app.Lock()
	defer app.Unlock()
	return app.Config.Authkey
}

func (app *Application) StopSyncServer() {
	if app.Server != nil {
		app.Server.GracefulStop()
	}

}

func (app *Application) BeforeExit() {
	app.StopSyncServer()
	BeforeUnMount()
}

func (app *Application) Run() {
	go StartMount(app.Config)
	systray.Run(onReady, onExit)
}

var App *Application

func NewApplication(config *Config, execPath string) (*Application, error) {
	app := new(Application)
	var err error

	app.Config = config
	app.ExecPath = execPath
	App = app
	return app, err
}

func onReady() {
	systray.SetIcon(static.IconPngByte)
	systray.SetTitle("")
	systray.SetTooltip("")
	systray.SetOnClick(func(menu systray.IMenu) {
		menu.ShowMenu()
	})
	systray.SetOnRClick(func(menu systray.IMenu) {
		menu.ShowMenu()
	})

	{
		mi := systray.AddMenuItem("打开目录", "")
		mi.Click(func() {
			go utils.OpenFileManager(App.ExecPath)
		})
	}
	if App.IsDebug {
		mi := systray.AddMenuItem("打开Temp目录", "")
		mi.Click(func() {
			go utils.OpenFileManager(filepath.Join(os.TempDir(), iconst.MountFolder))
		})
	}

	{
		systray.AddSeparator()

		isAutoRun, _ := utils.QueryAutoRun()
		mi := systray.AddMenuItemCheckbox("开机启动", "", isAutoRun)
		mi.Click(func() {
			if mi.Checked() {
				if err := utils.DisableAutoRun(); err != nil {
					logger.Logger.Error(err)
				} else {
					mi.Uncheck()
				}
			} else {
				if err := utils.EnableAutoRun(); err != nil {
					logger.Logger.Error(err)
				} else {
					mi.Check()
				}
			}
		})
	}
	{
		mi := systray.AddMenuItem("Github", "")
		mi.Click(func() {
			utils.OpenBrowser("https://github.com/jqwel/xlipboard")
		})
	}

	{
		systray.AddSeparator()
		systray.AddMenuItem("退出", "").Click(func() {
			systray.Quit()
		})
	}
}

func onExit() {
	tempFileFolder := filepath.Join(os.TempDir(), iconst.PngFolder)
	if err := os.RemoveAll(filepath.ToSlash(tempFileFolder)); err != nil {
		logger.Logger.Errorln(err)
	}
	os.Exit(0)
}
