package utils

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/jqwel/xlipboard/src/utils/logger"
)

type XlipboardService struct {
}

func (c *XlipboardService) ClipboardSequence() (string, error) {
	return fmt.Sprintf("%d", GetFixedNow().Nanosecond()), errors.New("")
}

func (c *XlipboardService) ContentType() (string, error) {
	fileResult, err := c.tryGetFiles()
	if err != nil {
		return "", err
	}
	if len(fileResult) > 0 {
		return TypeFile, nil
	}
	return TypeText, nil
}

func (c *XlipboardService) tryGetFiles() ([]string, error) {
	script := `
		use framework "AppKit"
		property this : a reference to current application
		property NSPasteboard : a reference to NSPasteboard of this
		property NSURL : a reference to NSURL of this
		property text item delimiters : linefeed

		set pb to NSPasteboard's generalPasteboard()
		set fs to (pb's readObjectsForClasses:[NSURL] options:[]) as list

		repeat with f in fs
			set f's contents to POSIX path of f
		end repeat

		fs as text
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// 使用osascript执行AppleScript命令
	cmd := exec.CommandContext(ctx, "osascript", "-e", script)

	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		logger.Logger.Error("Error executing command:", err)
		return nil, err
	}
	var result []string
	for _, v := range strings.Split(out.String(), "\n") {
		if v != "" {
			result = append(result, v)
		}
	}
	return result, nil
}

func (c *XlipboardService) Bitmap() ([]byte, error) {
	return nil, errors.New("TODO")
}

func (c *XlipboardService) Files() (filenames []string, err error) {
	files, err := c.tryGetFiles()
	return files, err
}

func (c *XlipboardService) SetFiles(paths []string) error {
	var onePath string
	if len(paths) == 0 {
		return nil
	} else if len(paths) == 1 {
		onePath = paths[0]
	} else {
		onePath = filepath.Dir(paths[0])
	}
	script := fmt.Sprintf(`tell app "Finder" to set the clipboard to ( POSIX file "%s" )`, onePath)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "osascript", "-e", script)
	err := cmd.Run()
	if err != nil {
		logger.Logger.Error("Error executing command:", err)
		return err
	}
	return nil
}
