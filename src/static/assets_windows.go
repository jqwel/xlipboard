package static

import (
	"bytes"
	_ "embed"
	"fmt"
	"image/png"

	ico "github.com/Kodeworks/golang-image-ico"
)

//go:embed design64.png
var IconPngByte0 []byte

var IconPngByte []byte

func init() {
	img, err := png.Decode(bytes.NewReader(IconPngByte0))
	if err != nil {
		fmt.Println("Error decoding PNG image:", err)
		return
	}

	buffer := new(bytes.Buffer)
	err = ico.Encode(buffer, img)
	if err != nil {
		fmt.Println("ico.Encode:", err)
		return
	}
	IconPngByte = buffer.Bytes()
}
