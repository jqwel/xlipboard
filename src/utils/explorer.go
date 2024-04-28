package utils

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/jqwel/xlipboard/src/utils/logger"
)

func OpenFileManager(dir string) {
	osType := runtime.GOOS
	var cmd *exec.Cmd
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	switch osType {
	case "windows":
		if dir == "" {
			dir = os.Getenv("USERPROFILE")
		}
		cmd = exec.CommandContext(ctx, "explorer", dir)
	case "darwin":
		if dir == "" {
			dir = os.Getenv("HOME")
		}
		cmd = exec.CommandContext(ctx, "open", dir)
	case "linux":
		if dir == "" {
			dir = os.Getenv("HOME")
		}
		cmd = exec.CommandContext(ctx, "xdg-open", dir)
	default:
		logger.Logger.Errorln("Unsupported operating system")
		return
	}
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error opening file manager:", err)
	}
}
