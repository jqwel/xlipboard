//go:build !no_fuse
// +build !no_fuse

package tags

func NoFuse() bool {
	return false
}
