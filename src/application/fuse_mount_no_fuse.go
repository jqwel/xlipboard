//go:build no_fuse
// +build no_fuse

package application

import (
	"github.com/jqwel/xlipboard/src/utils/logger"
)

func StartMount(c *Config) {
	logger.Logger.Println("NO_FUSE")
}

func BeforeUnMount() {
	logger.Logger.Println("NO_FUSE")
}
