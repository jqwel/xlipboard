package utils

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
)

const desktopFilePath = ".config/autostart/xlipboard.desktop"

func queryAutoRun() (bool, error) {
	usr, err := user.Current()
	if err != nil {
		return false, err
	}
	desktopPath := filepath.Join(usr.HomeDir, desktopFilePath)
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
	desktopPath := filepath.Join(usr.HomeDir, desktopFilePath)
	desktopContent := fmt.Sprintf(`[Desktop Entry]
Type=Application
Name=Xlipboard
Exec=%s
StartupNotify=false`, REG_VALUE)

	return os.WriteFile(desktopPath, []byte(desktopContent), 0644)
}

func disableAutoRun() error {
	usr, err := user.Current()
	if err != nil {
		return err
	}
	desktopPath := filepath.Join(usr.HomeDir, desktopFilePath)
	return os.Remove(desktopPath)
}
