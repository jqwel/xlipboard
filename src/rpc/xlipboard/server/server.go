package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/jqwel/xlipboard/src/tags"
	"image"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"golang.org/x/image/bmp"

	"github.com/jqwel/xlipboard/src/utils/logger"

	"github.com/jqwel/xlipboard/src/application"
	"github.com/jqwel/xlipboard/src/rpc/iconst"
	"github.com/jqwel/xlipboard/src/rpc/service"
	"github.com/jqwel/xlipboard/src/utils"
)

var ServerInstance = &Server{}

type Server struct {
	service.UnimplementedSyncServiceServer
	sm sync.Map
}

func (s *Server) Ping(ctx context.Context, req *service.PingRequest) (*service.PingReply, error) {
	return &service.PingReply{
		Msg: "pong",
	}, nil
}

func (s *Server) SayHello(ctx context.Context, req *service.HelloRequest) (*service.HelloReply, error) {
	return &service.HelloReply{
		Message:   iconst.Success,
		Timestamp: application.App.GetChangeAt(),
		Virtual:   application.App.GetVirtual(),
		Now:       utils.GetFixedNow().UnixMilli(),
	}, nil
}

func (s *Server) SayHowAreYou(ctx context.Context, req *service.HowAreYouRequest) (*service.HowAreYouReply, error) {
	localChangeAt := application.App.GetChangeAt()
	contentType, err := utils.Xlipboard().ContentType()
	if err != nil || contentType == utils.TypeBitmap || tags.NoFuse() {
		contentType = utils.TypeText
	}
	var copyStr string
	var picBytes []byte
	var copyFilename []string
	if contentType == utils.TypeText {
		copyStr, err = utils.Xlipboard().Text()
		if err != nil {
			return nil, err
		}
	} else if contentType == utils.TypeBitmap {
		picBytes, err = utils.Xlipboard().Bitmap()
		if err != nil {
			return nil, err
		}
		bmpBytesReader := bytes.NewReader(picBytes)
		var bmpImage image.Image
		bmpImage, err = bmp.Decode(bmpBytesReader)

		png0BytesReader := bytes.NewReader(picBytes)
		png0Image, err0 := png.Decode(png0BytesReader)
		if err != nil {
			if err0 != nil {
				return nil, err
			}
			err = nil
			bmpImage = png0Image // use png image
		}
		pngBytesBuffer := new(bytes.Buffer)
		err = png.Encode(pngBytesBuffer, bmpImage)
		if err != nil {
			return nil, err
		}
		copyImageByte := pngBytesBuffer.Bytes()
		picBytes = copyImageByte
	} else if contentType == utils.TypeFile {
		copyFilename, err = utils.Xlipboard().Files()
		sort.Strings(copyFilename)
	} else {
		return nil, errors.New("unsupported content type")
	}
	if contentType == utils.TypeBitmap {
		var XlipFsBytes = func(changeAt int64, contentType string, pngBytes []byte) ([]string, error) {
			tempDir := os.TempDir()
			tempFilePath := filepath.ToSlash(filepath.Join(tempDir, fmt.Sprintf("%s/xlipboard_%d.png", iconst.PngFolder, changeAt)))
			err := os.MkdirAll(filepath.Join(tempDir, iconst.PngFolder), 0755)
			if err != nil {
				return nil, err
			}
			err = os.WriteFile(tempFilePath, pngBytes, 0644)
			if err != nil {
				logger.Logger.Errorln("写入临时文件时发生错误：", err)
				return nil, err
			}
			return []string{tempFilePath}, nil
		}
		filenames, err := XlipFsBytes(localChangeAt, contentType, picBytes)
		if err != nil {
			return nil, err
		}
		if filenames != nil {
			contentType = utils.TypeFile
			copyFilename = filenames
			picBytes = nil
		}
	}
	for i := range copyFilename {
		copyFilename[i] = filepath.ToSlash(copyFilename[i])
	}
	return &service.HowAreYouReply{
		Message:       iconst.Success,
		Timestamp:     localChangeAt,
		ContentType:   contentType,
		CopyStr:       copyStr,
		CopyImageByte: picBytes,
		CopyFilename:  copyFilename,
	}, nil
}

func (s *Server) SetFileFD(fileMap *FileMap, fd uint64) {
	s.sm.Store(fd, fileMap)
}
func (s *Server) GetFileByFD(fd uint64) *FileMap {
	value, ok := s.sm.Load(fd)
	if !ok {
		logger.Logger.Errorf("GetFileByFD fd = %d Not Found", fd)
		return nil
	}
	return value.(*FileMap)
}
func (s *Server) ReleaseFileByFD(fd uint64) {
	byFD := s.GetFileByFD(fd)
	if byFD == nil {
		return
	}
	go func() {
		time.Sleep(time.Minute * 15)
		s.sm.Delete(fd)
		byFD.MemClear()
	}()
	byFD.Clear()
}

func (s *Server) Open(ctx context.Context, req *service.OpenRequest) (*service.OpenReply, error) {
	fileMap := NewFileMap(req.GetPath())
	fh := uint64(utils.GetFixedNow().UnixMicro())
	s.SetFileFD(fileMap, fh)
	return &service.OpenReply{
		Message: iconst.Success,
		Fh:      fh,
	}, nil
}
func (s *Server) Release(ctx context.Context, req *service.ReleaseRequest) (*service.ReleaseReply, error) {
	go s.ReleaseFileByFD(req.GetFh())
	return &service.ReleaseReply{
		Message: iconst.Success,
	}, nil
}

func (s *Server) ReadFile(ctx context.Context, req *service.ReadFileRequest) (*service.ReadFileReply, error) {
	fileMap := s.GetFileByFD(req.GetFh())
	if fileMap == nil {
		logger.Logger.Debug("ReadFile fileMap == nil")
		file, err := os.Open(req.GetPath())
		if err != nil {
			return nil, err
		}
		defer file.Close()
		_, err = file.Seek(req.GetOffset(), io.SeekStart)
		if err != nil {
			return nil, err
		}
		localBuf := make([]byte, req.GetLenBuf())
		n, err := file.Read(localBuf)
		if err != nil && err != io.EOF {
			return nil, err
		}
		return &service.ReadFileReply{
			Message: iconst.Success,
			Buf:     localBuf[:n],
			Read:    int32(n),
			EOF:     err == io.EOF,
		}, nil
	}
	if fileMap == nil {
		return nil, errors.New("ReadFile File Not Found")
	}
	file, ok, err := fileMap.GetOne()
	if err != nil {
		return nil, err
	}
	for !ok {
		time.Sleep(time.Millisecond * 100)
		file, ok, err = fileMap.GetOne()
		if err != nil {
			return nil, err
		}
	}
	var put = false
	var fnPutBack = func() {
		if !put {
			go fileMap.PutBack(file)
			put = true
		}
	}
	defer fnPutBack()

	_, err = file.Seek(req.GetOffset(), io.SeekStart)
	if err != nil {
		return nil, err
	}

	localBuf := make([]byte, req.GetLenBuf()*req.GetMul())
	n, err := file.Read(localBuf)
	fnPutBack()
	if err != nil && err != io.EOF {
		return nil, err
	}
	return &service.ReadFileReply{
		Message: iconst.Success,
		Buf:     localBuf[:n],
		Read:    int32(n),
		EOF:     err == io.EOF,
	}, nil
}

func (s *Server) ReadFileStream(req *service.ReadFileRequest, stream service.SyncService_ReadFileStreamServer) error {
	fileMap := s.GetFileByFD(req.GetFh())
	if fileMap == nil {
		return errors.New("ReadFileStream File Not Found")
	}
	file, ok, err := fileMap.GetOne()
	if err != nil {
		return err
	}
	for !ok {
		time.Sleep(time.Millisecond * 200)
		file, ok, err = fileMap.GetOne()
		if err != nil {
			return err
		}
	}
	var put = false
	var fnPutBack = func() {
		if !put {
			fileMap.PutBack(file)
			put = true
		}
	}
	defer fnPutBack()

	_, err = file.Seek(req.GetOffset(), io.SeekStart)
	if err != nil {
		return err
	}

	localBuf := make([]byte, req.GetLenBuf())
	n, err := file.Read(localBuf)
	fnPutBack()
	if err != nil && err != io.EOF {
		return err
	}
	if err == io.EOF || n == 0 {
		return stream.Send(&service.ReadFileReply{
			Message: iconst.Success,
			Buf:     nil,
			Read:    0,
			EOF:     true,
		})
	}

	const STREAM_SIZE = 4 * 1024
	for i := 0; i < n; i += STREAM_SIZE {
		j := i + STREAM_SIZE
		if j > n {
			j = n
		}
		err = stream.Send(&service.ReadFileReply{
			Message: iconst.Success,
			Buf:     localBuf[i:j],
			Read:    int32(j - i),
			EOF:     false,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) ReadDir(ctx context.Context, req *service.ReadDirRequest) (*service.ReadDirReply, error) {
	path := req.GetPath()
	dir, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var dirInfoList = make([]*service.ReadDirReply_DirInfo, 0)
	for i := range dir {
		fileInfo, err := os.Stat(path + "/" + dir[i].Name())
		if err != nil {
			return nil, err
		}

		var lastAccessTime, statusChangeTime, birthTime, flags = AdditionalFileAttribute(fileInfo)

		dirInfoList = append(dirInfoList, &service.ReadDirReply_DirInfo{
			Name:             fileInfo.Name(),
			IsDir:            fileInfo.IsDir(),
			FileMode:         uint32(fileInfo.Mode()),
			Size:             fileInfo.Size(),
			ModTime:          fileInfo.ModTime().UnixMilli(),
			AccessTime:       lastAccessTime,
			StatusChangeTime: statusChangeTime,
			BirthTime:        birthTime,
			Flags:            flags,
		})
	}
	return &service.ReadDirReply{
		Message:     iconst.Success,
		DirInfoList: dirInfoList,
	}, nil
}

func (s *Server) Stat(ctx context.Context, req *service.StatRequest) (*service.StatReply, error) {
	path := req.GetPath()
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	var lastAccessTime, statusChangeTime, birthTime, flags = AdditionalFileAttribute(fileInfo)

	return &service.StatReply{
		Message:          iconst.Success,
		Name:             fileInfo.Name(),
		IsDir:            fileInfo.IsDir(),
		FileMode:         uint32(fileInfo.Mode()),
		Size:             fileInfo.Size(),
		ModTime:          fileInfo.ModTime().UnixMilli(),
		AccessTime:       lastAccessTime,
		StatusChangeTime: statusChangeTime,
		BirthTime:        birthTime,
		Flags:            flags,
	}, nil
}
