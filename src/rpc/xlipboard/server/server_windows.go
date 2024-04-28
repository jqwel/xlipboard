package server

import (
	"os"
	"syscall"
)

func AdditionalFileAttribute(fileInfo os.FileInfo) (int64, int64, int64, uint32) {
	var lastAccessTime int64
	var statusChangeTime int64
	var birthTime int64
	var flags uint32
	if stat, ok := fileInfo.Sys().(*syscall.Win32FileAttributeData); ok {
		lastAccessTime = stat.LastAccessTime.Nanoseconds() / 1e6
		statusChangeTime = 0
		birthTime = stat.CreationTime.Nanoseconds() / 1e6
		flags = stat.FileAttributes
	}
	return lastAccessTime, statusChangeTime, birthTime, flags
}
