package utils

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

func getDesktopFilePath() {
	return fmt.Sprintf(".config/autostart/%s.desktop", strings.ToLower(REG_KEY))
}

func queryAutoRun() (bool, error) {
	usr, err := user.Current()
	if err != nil {
		return false, err
	}
	desktopPath := filepath.Join(usr.HomeDir, getDesktopFilePath())
	_, err = os.Stat(desktopPath)
	if err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func enableAutoRun() error {
	usr, err := user.Current()
	if err != nil {
		return err
	}
	desktopPath := filepath.Join(usr.HomeDir, getDesktopFilePath())
	desktopContent := fmt.Sprintf(`[Desktop Entry]
Type=Application
Name=%s
Exec=%s
StartupNotify=false`, REG_KEY, REG_VALUE)

	return os.WriteFile(desktopPath, []byte(desktopContent), 0644)
}

func disableAutoRun() error {
	usr, err := user.Current()
	if err != nil {
		return err
	}
	desktopPath := filepath.Join(usr.HomeDir, getDesktopFilePath())
	return os.Remove(desktopPath)
}
