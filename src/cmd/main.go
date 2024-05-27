package main

import (
	"os"
	"path/filepath"
	"time"

	"github.com/jqwel/xlipboard/src/rpc/xlipboard/rpc_app"
	"github.com/jqwel/xlipboard/src/utils/logger"

	"github.com/sirupsen/logrus"

	"github.com/jqwel/xlipboard/src/application"
	"github.com/jqwel/xlipboard/src/utils"
)

var app *application.Application

var execPath string
var execFullPath string
var config *application.Config
var mode = "debug"
var version = ""
var log = logger.Logger

func init() {
	execFullPath = os.Args[0]
	execPath = filepath.Dir(execFullPath)

	var err error
	configFilePath := filepath.ToSlash(filepath.Join(execPath, application.ConfigFile))
	config, err = application.LoadConfig(configFilePath)
	if err != nil {
		log.WithError(err).Warn("failed to load config")
	}
	log.SetLevel(logrus.DebugLevel)

	if mode == "debug" {
		log.SetLevel(logrus.DebugLevel)
		log.SetFormatter(&logrus.TextFormatter{ForceColors: true})
	} else {
		log.SetLevel(logrus.FatalLevel)
		//log.SetLevel(logrus.ErrorLevel)
		//logFilePath := filepath.ToSlash(filepath.Join(execPath, application.LogFile))
		//f, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND, 0644)
		//if err != nil {
		//	log.WithError(err).Fatal("failed to open log file")
		//}
		//log.SetOutput(f)
	}
}

func main() {
	var err error

	app, err = application.NewApplication(config, execPath)
	if err != nil {
		log.WithError(err).Fatal("failed to create applicaton")
	}
	defer app.BeforeExit()
	app.IsDebug = mode == "debug"

	if err := utils.InitOffsetPeriodically(app.Config.NtpAddress, time.Minute*10); err != nil {
		logger.Logger.Error(err)
	}

	go application.NewDetector().Start()
	log.Debug("start server")
	go func() {
		//if err := rpc_app.RunGrpcServer(app.Config.Certificate, app.Config.PrivateKey); err != nil {
		//	logger.Logger.Error(err)
		//	app.BeforeExit()
		//}
	}()

	go func() {
		if err := rpc_app.RunQuicServer(app.Config.Certificate, app.Config.PrivateKey); err != nil {
			logger.Logger.Error(err)
			app.BeforeExit()
		}
	}()

	log.Debug("start app")
	app.Run()
}
