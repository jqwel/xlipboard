package utils

import (
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
	"syscall"
	"unsafe"

	"github.com/lxn/walk"
	"github.com/lxn/win"
	"golang.org/x/sys/windows"
)

var (
	user32                     = syscall.MustLoadDLL("user32")
	getClipboardSequenceNumber = user32.MustFindProc("GetClipboardSequenceNumber")
)

var Formats = []uint32{win.CF_HDROP, win.CF_DIBV5, win.CF_UNICODETEXT}

type XlipboardService struct {
	hwnd                     win.HWND
	contentsChangedPublisher walk.EventPublisher
}

func (c *XlipboardService) ClipboardSequence() (string, error) {
	r, _, _ := getClipboardSequenceNumber.Call()
	return fmt.Sprintf("%d", r), nil
}

func (c *XlipboardService) ContentType() (string, error) {
	var format uint32
	err := c.withOpenClipboard(func() error {
		for _, f := range Formats {
			isAvaliable := win.IsClipboardFormatAvailable(f)
			if isAvaliable {
				format = f
				return nil
			}
		}
		return lastError("get content type of clipboard")
	})
	if err != nil {
		return "", err
	}
	switch format {
	case win.CF_HDROP:
		return TypeFile, nil
	case win.CF_DIBV5:
		return TypeBitmap, nil
	case win.CF_UNICODETEXT:
		return TypeText, nil
	default:
		return TypeUnknown, nil
	}
}

func int32Abs(val int32) uint32 {
	if val < 0 {
		return uint32(-val)
	}
	return uint32(val)
}

func (c *XlipboardService) Bitmap() (bmpBytes []byte, err error) {
	err = c.withOpenClipboard(func() error {
		hMem := win.HGLOBAL(win.GetClipboardData(win.CF_DIBV5))
		if hMem == 0 {
			return lastError("GetClipboardData")
		}

		p := win.GlobalLock(hMem)
		if p == nil {
			return lastError("GlobalLock()")
		}
		defer win.GlobalUnlock(hMem)

		header := (*win.BITMAPV5HEADER)(unsafe.Pointer(p))
		var biSizeImage uint32
		if header.BiBitCount == 32 {
			biSizeImage = 4 * int32Abs(header.BiWidth) * int32Abs(header.BiHeight)
		} else {
			biSizeImage = header.BiSizeImage
		}

		var data []byte
		sh := (*reflect.SliceHeader)(unsafe.Pointer(&data))
		sh.Data = uintptr(p)
		sh.Cap = int(header.BiSize + biSizeImage)
		sh.Len = int(header.BiSize + biSizeImage)

		if header.BiCompression == 3 && header.BV4RedMask == 0xff0000 && header.BV4GreenMask == 0xff00 && header.BV4BlueMask == 0xff {
			header.BiCompression = win.BI_RGB
			pixelStartAt := header.BiSize
			for i := pixelStartAt + 3; i < uint32(len(data)); i += 4 {
				data[i] = 0xff
			}
		}

		bmpFileSize := 14 + header.BiSize + biSizeImage
		bmpBytes = make([]byte, bmpFileSize)

		binary.LittleEndian.PutUint16(bmpBytes[0:], 0x4d42) // start with 'BM'
		binary.LittleEndian.PutUint32(bmpBytes[2:], bmpFileSize)
		binary.LittleEndian.PutUint16(bmpBytes[6:], 0)
		binary.LittleEndian.PutUint16(bmpBytes[8:], 0)
		binary.LittleEndian.PutUint32(bmpBytes[10:], 14+header.BiSize)
		copy(bmpBytes[14:], data[:])

		return nil
	})
	return
}

func (c *XlipboardService) Files() (filenames []string, err error) {
	err = c.withOpenClipboard(func() error {
		hMem := win.HGLOBAL(win.GetClipboardData(win.CF_HDROP))
		if hMem == 0 {
			return lastError("GetClipboardData")
		}
		p := win.GlobalLock(hMem)
		if p == nil {
			return lastError("GlobalLock()")
		}
		defer win.GlobalUnlock(hMem)
		filesCount := win.DragQueryFile(win.HDROP(p), 0xFFFFFFFF, nil, 0)
		filenames = make([]string, 0, filesCount)
		buf := make([]uint16, win.MAX_PATH)
		for i := uint(0); i < filesCount; i++ {
			win.DragQueryFile(win.HDROP(p), i, &buf[0], win.MAX_PATH)
			filenames = append(filenames, windows.UTF16ToString(buf))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return
}

type DROPFILES struct {
	pFiles uintptr
	pt     uintptr
	fNC    bool
	fWide  bool
	_      uint32 // padding
}

// SetFiles sets the current file drop data of the clipboard.
func (c *XlipboardService) SetFiles(paths []string) error {
	return c.withOpenClipboard(func() error {
		win.EmptyClipboard()
		// https://docs.microsoft.com/en-us/windows/win32/shell/clipboard#cf_hdrop
		var utf16 []uint16
		for _, path := range paths {
			_utf16, err := syscall.UTF16FromString(path)
			if err != nil {
				return err
			}
			utf16 = append(utf16, _utf16...)
		}
		utf16 = append(utf16, uint16(0))

		const dropFilesSize = unsafe.Sizeof(DROPFILES{}) - 4

		size := dropFilesSize + uintptr((len(utf16))*2+2)

		hMem := win.GlobalAlloc(win.GHND, size)
		if hMem == 0 {
			return lastError("GlobalAlloc")
		}

		p := win.GlobalLock(hMem)
		if p == nil {
			return lastError("GlobalLock()")
		}

		zeroMem := make([]byte, size)
		win.MoveMemory(p, unsafe.Pointer(&zeroMem[0]), size)

		pD := (*DROPFILES)(p)
		pD.pFiles = dropFilesSize
		pD.fWide = false
		pD.fNC = true
		win.MoveMemory(unsafe.Pointer(uintptr(p)+dropFilesSize), unsafe.Pointer(&utf16[0]), uintptr(len(utf16)*2))

		win.GlobalUnlock(hMem)

		if 0 == win.SetClipboardData(win.CF_HDROP, win.HANDLE(hMem)) {
			defer win.GlobalFree(hMem)

			return lastError("SetClipboardData")
		}

		return nil
	})
}

func (c *XlipboardService) withOpenClipboard(f func() error) error {
	if !win.OpenClipboard(c.hwnd) {
		return lastError("OpenClipboard")
	}
	defer win.CloseClipboard()

	return f()
}

func lastError(name string) error {
	return errors.New(fmt.Sprintf("%s failed", name))
}
