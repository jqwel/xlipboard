package utils

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
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
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "xclip", "-o", "-selection", "clipboard", "-t", "TARGETS")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		logger.Logger.Errorln("Error executing command:", err)
		return "", err
	}
	output := strings.TrimSpace(out.String())
	typeSlice := strings.Split(output, "\n")
	if InSlice("x-special/gnome-copied-files", typeSlice) {
		return TypeFile, nil
	} else if InSlice("UTF8_STRING", typeSlice) {
		return TypeText, nil
	} else if InSlice("image/png", typeSlice) {
		return TypeBitmap, nil
	} else {
		return TypeUnknown, nil
	}
}

func (c *XlipboardService) Bitmap() ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "xclip", "-o", "-selection", "clipboard", "-t", "image/png")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		logger.Logger.Errorln("Error executing command:", err)
		return nil, err
	}

	data := make([]byte, out.Len())
	read, err := out.Read(data)
	if err != nil {
		return nil, err
	}
	if read != len(data) {
		return nil, errors.New(`read != out.Len()`)
	}
	return data, nil
}

func (c *XlipboardService) Files() (filenames []string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "xclip", "-o", "-selection", "clipboard")
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		logger.Logger.Errorln("Error executing command:", err)
		return nil, err
	}
	output := strings.TrimSpace(out.String())
	filenamesOpt := strings.Split(output, "\n")

	for _, filename := range filenamesOpt {
		if filename == "copy" || filename == "" {
			continue
		}
		if strings.Index(filename, "file://") == 0 {
			fixed := strings.Replace(filename, "file://", "", 1)
			filenames = append(filenames, fixed)
		} else {
			filenames = append(filenames, filename)
		}
	}
	return
}

func (c *XlipboardService) SetFiles(paths []string) error {
	if len(paths) == 0 {
		return nil
	}
	input := "copy"
	for _, filepath := range paths {
		input += "\nfile://" + filepath
	}
	input = input + "\\0"
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	cmd0 := exec.CommandContext(ctx, "bash", "-c", `echo -e "`+input+`" | xclip -i -selection clipboard -t x-special/gnome-copied-files`)
	return cmd0.Run()
}
