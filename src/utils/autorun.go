package utils

import (
	"os"
)

const REG_KEY = "Xlipboard"

var REG_VALUE = os.Args[0]

func QueryAutoRun() (bool, error) {
	return queryAutoRun()
}
func EnableAutoRun() error {
	return enableAutoRun()
}
func DisableAutoRun() error {
	return disableAutoRun()
}
