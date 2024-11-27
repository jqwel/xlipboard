package utils

import (
	"os"
)

var REG_KEY = "Xlipboard"

var REG_VALUE = os.Args[0]

func SetRegKey(key string) {
	REG_KEY = key
}
func QueryAutoRun() (bool, error) {
	return queryAutoRun()
}
func EnableAutoRun() error {
	return enableAutoRun()
}
func DisableAutoRun() error {
	return disableAutoRun()
}
