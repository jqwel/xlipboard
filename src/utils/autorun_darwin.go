package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

const launchAgentPath = "Library/LaunchAgents/github.jqwel.xlipboard.plist"

var launchAgentContent = fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>github.jqwel.xlipboard</string>
	<key>ProgramArguments</key>
	<array>
		<string>%s</string>
	</array>
	<key>RunAtLoad</key>
	<true/>
</dict>
</plist>
`, REG_VALUE)

func queryAutoRun() (bool, error) {
	_, err := os.Stat(filepath.Join(os.Getenv("HOME"), launchAgentPath))
	if err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func enableAutoRun() error {
	launchAgent := filepath.Join(os.Getenv("HOME"), launchAgentPath)
	return os.WriteFile(launchAgent, []byte(launchAgentContent), 0644)
}

func disableAutoRun() error {
	launchAgent := filepath.Join(os.Getenv("HOME"), launchAgentPath)
	return os.Remove(launchAgent)
}
