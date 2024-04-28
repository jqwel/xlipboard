package server

import (
	"os"
	"syscall"
)

func AdditionalFileAttribute(fileInfo os.FileInfo) (int64, int64, int64, uint32) {
	var lastAccessTime int64
	var statusChangeTime int64
	var birthTime int64
	var permissions uint32

	if stat, ok := fileInfo.Sys().(*syscall.Stat_t); ok {
		lastAccessTime = stat.Atimespec.Nano() / 1e6
		statusChangeTime = stat.Ctimespec.Nano() / 1e6
		birthTime = stat.Birthtimespec.Nano() / 1e6
		permissions = stat.Flags
	}
	return lastAccessTime, statusChangeTime, birthTime, permissions
}
