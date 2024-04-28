package client

import (
	"time"

	"github.com/jqwel/xlipboard/src/rpc/iconst"
	"github.com/jqwel/xlipboard/src/rpc/service"
	"github.com/jqwel/xlipboard/src/utils/logger"
)

const MaxRetry = iconst.MaxLenChanConn * 10

func SayHello(target string, currentChangeAt int64) (*service.HelloReply, error) {
	for i := range 1 {
		if g, err := SayHelloQ(target, currentChangeAt); err != nil {
			logger.Logger.WithError(err).Error("SayHelloG Error")
		} else {
			if i > 0 {
				logger.Logger.WithField("i", i).Info("SayHelloG Retry")
			}
			return g, err
		}
	}
	return SayHelloQ(target, currentChangeAt)
}

func SayHowAreYou(target string, forChangeAt int64) (*service.HowAreYouReply, error) {
	for i := range 1 {
		if g, err := SayHowAreYouQ(target, forChangeAt); err != nil {
			logger.Logger.WithError(err).Error("SayHowAreYouG Error")
		} else {
			if i > 0 {
				logger.Logger.WithField("i", i).Info("SayHowAreYouG Retry")
			}
			return g, err
		}
	}
	return SayHowAreYouQ(target, forChangeAt)
}

func ReadDir(target string, path string) (*service.ReadDirReply, error) {
	for i := range 1 {
		if g, err := ReadDirQ(target, path); err != nil {
			logger.Logger.WithError(err).Error("ReadDirG Error")
		} else {
			if i > 0 {
				logger.Logger.WithField("i", i).Info("ReadDirG Retry")
			}
			return g, err
		}
	}
	return ReadDirQ(target, path)
}

func Open(target string, path string) (*service.OpenReply, error) {
	for i := range 1 {
		if g, err := OpenQ(target, path); err != nil {
			logger.Logger.WithError(err).Error("OpenG Error")
		} else {
			if i > 0 {
				logger.Logger.WithField("i", i).Info("OpenG Retry")
			}
			return g, err
		}
	}
	return OpenQ(target, path)
}

func Release(target string, path string, fh uint64) (*service.ReleaseReply, error) {
	for i := range 1 {
		if g, err := ReleaseQ(target, path, fh); err != nil {
			logger.Logger.WithError(err).Error("ReleaseG Error")
		} else {
			if i > 0 {
				logger.Logger.WithField("i", i).Info("ReleaseG Retry")
			}
			return g, err
		}
	}
	return ReleaseQ(target, path, fh)
}

func ReadFile(target string, path string, buf []byte, offset int64, fh uint64, mul uint32) (*service.ReadFileReply, error) {
	for i := range MaxRetry {
		if g, err := ReadFileQ(target, path, buf, offset, fh, mul); err != nil {
			logger.Logger.WithError(err).Error("ReadFileG Error")
			if i > MaxRetry/2 {
				time.Sleep(time.Second)
			}
		} else {
			if i > 0 {
				logger.Logger.WithField("i", i).Info("ReadFileG Retry")
			}
			return g, err
		}
	}
	return ReadFileQ(target, path, buf, offset, fh, mul)
}

func ReadFileStream(target string, path string, buf []byte, offset int64, fh uint64, mul uint32) (*service.ReadFileReply, error) {
	return ReadFileStreamG(target, path, buf, offset, fh, mul)
}

func Stat(target string, path string) (*service.StatReply, error) {
	for i := range 1 {
		if g, err := StatQ(target, path); err != nil {
			logger.Logger.WithError(err).Error("StatG Error")
		} else {
			if i > 0 {
				logger.Logger.WithField("i", i).Info("StatG Retry")
			}
			return g, err
		}
	}
	return StatQ(target, path)
}
