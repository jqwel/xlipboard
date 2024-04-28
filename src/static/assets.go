//go:build !windows

package static

import _ "embed"

//go:embed design64.png
var IconPngByte []byte
