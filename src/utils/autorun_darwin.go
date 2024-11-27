package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func getLaunchAgentPath() string {
	return fmt.Sprintf("Library/LaunchAgents/github.jqwel.%s.plist", strings.ToLower(REG_KEY))
}

func getLaunchAgentContent() string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>github.jqwel.%s</string>
	<key>ProgramArguments</key>
	<array>
		<string>%s</string>
	</array>
	<key>RunAtLoad</key>
	<true/>
</dict>
</plist>
`, strings.ToLower(REG_KEY), REG_VALUE)
}

func queryAutoRun() (bool, error) {
	_, err := os.Stat(filepath.Join(os.Getenv("HOME"), getLaunchAgentPath()))
	if err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func enableAutoRun() error {
	launchAgent := filepath.Join(os.Getenv("HOME"), getLaunchAgentPath())
	return os.WriteFile(launchAgent, []byte(getLaunchAgentContent()), 0644)
}

func disableAutoRun() error {
	launchAgent := filepath.Join(os.Getenv("HOME"), getLaunchAgentPath())
	return os.Remove(launchAgent)
}
