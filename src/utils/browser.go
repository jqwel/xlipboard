package utils

import (
	"context"
	"os/exec"
	"runtime"
	"time"

	"github.com/jqwel/xlipboard/src/utils/logger"
)

func OpenBrowser(url string) {
	var cmd *exec.Cmd

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	switch runtime.GOOS {
	case "windows":
		cmd = exec.CommandContext(ctx, "cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.CommandContext(ctx, "open", url)
	default:
		cmd = exec.CommandContext(ctx, "xdg-open", url)
	}

	if err := cmd.Run(); err != nil {
		logger.Logger.Error(err)
	}
}
