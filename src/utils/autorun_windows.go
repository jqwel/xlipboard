package utils

import (
	"golang.org/x/sys/windows/registry"
)

func queryAutoRun() (bool, error) {
	key, err := openAutoRunKey(registry.QUERY_VALUE)
	if err != nil {
		return false, err
	}
	defer key.Close()
	val, _, err := key.GetStringValue(REG_KEY)
	if err != nil {
		if err == registry.ErrNotExist {
			return false, nil
		}
		return false, err
	}
	if val == REG_VALUE {
		return true, nil
	}
	return false, nil
}

func openAutoRunKey(access uint32) (registry.Key, error) {
	autorunKey := `SOFTWARE\Microsoft\Windows\CurrentVersion\Run`
	key, err := registry.OpenKey(registry.CURRENT_USER, autorunKey, access)
	if err != nil {
		return 0, err
	}
	return key, nil
}

func enableAutoRun() error {
	key, err := openAutoRunKey(registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()
	return key.SetStringValue(REG_KEY, REG_VALUE)
}

func disableAutoRun() error {
	key, err := openAutoRunKey(registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()
	return key.DeleteValue(REG_KEY)
}
