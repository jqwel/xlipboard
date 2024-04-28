package application

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"github.com/winfsp/cgofuse/fuse"

	"github.com/jqwel/xlipboard/src/rpc/xlipboard/client"
)

var (
	_ fuse.FileSystemInterface = &XlipboardFs{}
)

type XlipboardFs struct {
	fuse.FileSystemBase
	XlipboardCache
}

func (gfs *XlipboardFs) Open(path string, flags int) (int, uint64) {
	if path == "/" {
		return 0, 0
	}
	list := GetFileList()
	path = filepath.ToSlash(path)
	split := strings.Split(path, "/")
	if len(split) < 1 {
		return -fuse.ENOENT, ^uint64(0)
	}
	timestampStr := split[1]
	var found *XlipFileRemote
	for i := range list {
		if fmt.Sprintf("%d", list[i].Timestamp) == timestampStr {
			found = list[i]
			break
		}
	}
	if found == nil || len(found.Filenames) == 0 {
		return -fuse.ENOENT, ^uint64(0)
	}

	if len(split) == 2 {
		return 0, 0
	}
	if len(split) == 3 {
		foundFile := false
		for i := range found.Filenames {
			if filepath.Base(found.Filenames[i]) == split[2] {
				foundFile = true
				break
			}
		}
		if !foundFile {
			return -fuse.ENOENT, ^uint64(0)
		}
	}
	pathRemote := filepath.ToSlash(filepath.Join(filepath.Dir(found.Filenames[0]), strings.Join(split[2:], "/")))
	reply, err := client.Open(found.Target, pathRemote)
	if err != nil {
		return -fuse.ENOENT, ^uint64(0)
	}
	gfs.InitCa(reply.GetFh())
	return 0, reply.GetFh()
}

func (gfs *XlipboardFs) Release(path string, fh uint64) int {
	gfs.ClearCa(fh)
	if path == "/" {
		return 0
	}
	list := GetFileList()
	path = filepath.ToSlash(path)
	split := strings.Split(path, "/")
	if len(split) < 1 {
		return 0
	}
	timestampStr := split[1]
	var found *XlipFileRemote
	for i := range list {
		if fmt.Sprintf("%d", list[i].Timestamp) == timestampStr {
			found = list[i]
			break
		}
	}
	if found == nil || len(found.Filenames) == 0 {
		return 0
	}

	if len(split) == 2 {
		return 0
	}
	if len(split) == 3 {
		foundFile := false
		for i := range found.Filenames {
			if filepath.Base(found.Filenames[i]) == split[2] {
				foundFile = true
				break
			}
		}
		if !foundFile {
			return 0
		}
	}
	pathRemote := filepath.ToSlash(filepath.Join(filepath.Dir(found.Filenames[0]), strings.Join(split[2:], "/")))
	_, err := client.Release(found.Target, pathRemote, fh)
	if err != nil {
		return -fuse.ENOENT
	}
	return 0
}

func (gfs *XlipboardFs) Getattr(path string, stat *fuse.Stat_t, fh uint64) int {
	if path == "/" {
		stat.Mode = fuse.S_IFDIR | 0555
		return 0
	}
	list := GetFileList()
	path = filepath.ToSlash(path)
	split := strings.Split(path, "/")
	if len(split) < 1 {
		return -fuse.ENOENT
	}
	if len(split) == 2 {
		stat.Mode = fuse.S_IFDIR | 0555
		return 0
	}
	timestampStr := split[1]
	var found *XlipFileRemote
	for i := range list {
		if fmt.Sprintf("%d", list[i].Timestamp) == timestampStr {
			found = list[i]
			break
		}
	}
	if found == nil || len(found.Filenames) == 0 {
		return -fuse.ENOENT
	}

	if len(split) == 3 {
		foundFile := false
		for i := range found.Filenames {
			if filepath.Base(found.Filenames[i]) == split[2] {
				foundFile = true
				break
			}
		}
		if !foundFile {
			return -fuse.ENOENT
		}
	}

	path0 := filepath.ToSlash(filepath.Join(filepath.Dir(found.Filenames[0]), strings.Join(split[2:], "/")))
	reply, err := client.Stat(found.Target, path0)
	if err != nil {
		return -fuse.ENOENT
	}

	stat.Mode = FileModeToFuseMode(fs.FileMode(reply.GetFileMode()), reply.GetIsDir())
	stat.Size = reply.GetSize()
	stat.Atim = Time64ToTimespec(reply.GetAccessTime())
	stat.Mtim = Time64ToTimespec(reply.GetModTime())
	stat.Ctim = Time64ToTimespec(reply.GetStatusChangeTime())
	stat.Birthtim = Time64ToTimespec(reply.GetBirthTime())
	return 0
}

func (gfs *XlipboardFs) Read(path string, buff []byte, ofstFix int64, fh uint64) int {
	ofst := ofstFix
	if ofst < 0 {
		ofst = -ofst
	}
	list := GetFileList()
	path = filepath.ToSlash(path)
	split := strings.Split(path, "/")
	if len(split) < 1 {
		return 0
	}
	timestampStr := split[1]
	var found *XlipFileRemote
	for i := range list {
		if fmt.Sprintf("%d", list[i].Timestamp) == timestampStr {
			found = list[i]
			break
		}
	}
	if found == nil || len(found.Filenames) == 0 {
		return 0
	}
	found.FsAccessed()

	pathRemote := filepath.ToSlash(filepath.Join(filepath.Dir(found.Filenames[0]), strings.Join(split[2:], "/")))
	if false {
		var mul uint32 = 3
		need := len(buff)
		requested := gfs.CheckCaRequested(fh, ofstFix, len(buff), mul)

		if !requested {
			if mul > 1 && ofstFix >= 0 {
				for i := range 2 {
					go func(i int) {
						_ = gfs.Read(path, make([]byte, len(buff)), -(ofst + int64(i+1)*int64(len(buff))*int64(mul)), fh)
					}(i)
				}
			}
			reply, err := client.ReadFile(found.Target, pathRemote, buff, ofst, fh, mul)
			if err != nil {
				return 0
			}
			read := reply.GetRead()
			if read > int32(need) {
				read = int32(need)
			}
			if ofstFix >= 0 {
				copy(buff, reply.GetBuf()[:read])
			}
			if read > 0 {
				gfs.ResultCaRequested(fh, ofst, len(buff), mul, reply.GetRead(), reply.GetBuf())
			}
			return int(read)
		} else {
			fetch := gfs.ResultCaFetch(fh, ofstFix, len(buff), mul)
			if ofstFix >= 0 && fetch != nil {
				copy(buff, fetch)
			}
			return len(fetch)
		}
	}
	if true {
		reply, err := client.ReadFile(found.Target, pathRemote, buff, ofst, fh, 1)
		if err != nil {
			return 0
		}
		copy(buff, reply.GetBuf()[:reply.GetRead()])
		return int(reply.GetRead())
	}
	return 0
}

func (gfs *XlipboardFs) Readdir(path string, fill func(name string, stat *fuse.Stat_t, ofst int64) bool, ofst int64, fh uint64) int {
	if ofst > 0 {
		return 0
	}
	fill(".", nil, ofst)
	fill("..", nil, ofst)
	filelist := GetFileList()
	if path == "/" {
		for i := range filelist {
			fi := filelist[i]
			fill(fmt.Sprintf("%d", fi.Timestamp), nil, ofst)
		}
	} else {
		split := strings.Split(path, "/")[1:]
		if len(split) >= 1 {
			var found *XlipFileRemote
			for i := range filelist {
				var xlipFileRemote = filelist[i]
				if fmt.Sprintf("%d", xlipFileRemote.Timestamp) == split[0] {
					found = xlipFileRemote
				}
			}

			if found != nil {
				filename := found.Filenames[0]
				absPath := filepath.Dir(filename)
				readpath := filepath.ToSlash(filepath.Join(absPath, strings.Join(split[1:], "/")))
				dir, err := client.ReadDir(found.Target, readpath)
				if err != nil {
					return -fuse.ENOENT
				}
				filterMap := make(map[string]bool)
				if len(split) < 2 {
					for i := range found.Filenames {
						filterMap[filepath.Base(found.Filenames[i])] = true
					}
				}
				for i := range dir.DirInfoList {
					dirInfo := dir.DirInfoList[i]
					if filterMap[dirInfo.GetName()] || len(split) >= 2 {
						fill(dirInfo.GetName(), &fuse.Stat_t{
							Mode:     dirInfo.GetFileMode(),
							Size:     dirInfo.GetSize(),
							Atim:     Time64ToTimespec(dirInfo.GetAccessTime()),
							Mtim:     Time64ToTimespec(dirInfo.GetModTime()),
							Ctim:     Time64ToTimespec(dirInfo.GetStatusChangeTime()),
							Birthtim: Time64ToTimespec(dirInfo.GetBirthTime()),
							Flags:    dirInfo.GetFlags(),
						}, ofst)
					}
				}
			}
		}
	}
	return 0
}

func Time64ToTimespec(timestamp64 int64) fuse.Timespec {
	return fuse.NewTimespec(time.Unix(timestamp64/1e3, (timestamp64%1000)*1000000))
}

func FileModeToFuseMode(mode fs.FileMode, isDir bool) uint32 {
	var result = uint32(mode & 0777)
	set := false
	if isDir || mode&fs.ModeDir == fs.ModeDir {
		result |= fuse.S_IFDIR
		set = true
	} else if mode&fs.ModeSymlink == fs.ModeSymlink {
		result |= fuse.S_IFLNK
		set = true
	} else if mode&fs.ModeNamedPipe == fs.ModeNamedPipe {
		result |= fuse.S_IFIFO
		set = true
	} else if mode&fs.ModeSocket == fs.ModeSocket {
		result |= fuse.S_IFSOCK
		set = true
	} else if mode&fs.ModeSetuid == fs.ModeSetuid {
		result |= fuse.S_ISUID
		set = true
	} else if mode&fs.ModeSetgid == fs.ModeSetgid {
		result |= fuse.S_ISGID
		set = true
	} else if mode&fs.ModeSticky == fs.ModeSticky {
		result |= fuse.S_ISVTX
		set = true
	} else if mode&fs.ModeCharDevice == fs.ModeCharDevice {
		result |= fuse.S_IFCHR
		set = true
	}
	if !set || mode&fs.ModeType == 0 {
		result |= fuse.S_IFREG
	}
	return result
}
