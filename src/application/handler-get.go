package application

import (
	"bytes"
	"errors"
	"image"
	"image/png"
	"sort"

	"github.com/jqwel/xlipboard/src/utils/logger"

	"golang.org/x/image/bmp"

	"github.com/jqwel/xlipboard/src/utils"
)

func getContentTypeAndData() (contentType string, copyStr string, copyImageByte []byte, copyFilename []string, err error) {
	contentType, err = utils.Xlipboard().ContentType()
	if err != nil {
		//log.WithError(err).Info("failed to get content type of clipboard")
		return
	}
	if contentType == utils.TypeBitmap {
		contentType = utils.TypeText // disable bitmap
	}
	if contentType == utils.TypeText {
		copyStr, err = utils.Xlipboard().Text()
		if len(copyStr) > 1*1024*1024 {
			copyStr = ""
		}
		if err != nil {
			logger.Logger.WithError(err).Warn("failed to get clipboard")
		}
		return
	} else if contentType == utils.TypeBitmap {
		var bmpBytes []byte
		bmpBytes, err = utils.Xlipboard().Bitmap()
		if err != nil {
			logger.Logger.WithError(err).Warn("failed to get bmp bytes from clipboard")
			return
		}

		bmpBytesReader := bytes.NewReader(bmpBytes)
		var bmpImage image.Image
		bmpImage, err = bmp.Decode(bmpBytesReader)

		png0BytesReader := bytes.NewReader(bmpBytes)
		png0Image, err0 := png.Decode(png0BytesReader)
		if err != nil {
			if err0 != nil {
				logger.Logger.WithError(err).Warn("failed to decode bmp")
				return
			}
			err = nil
			bmpImage = png0Image // use png image
		}
		pngBytesBuffer := new(bytes.Buffer)
		err = png.Encode(pngBytesBuffer, bmpImage)
		if err != nil {
			logger.Logger.WithError(err).Warn("failed to encode bmp as png")
			return
		}
		copyImageByte = pngBytesBuffer.Bytes()
		return
	} else if contentType == utils.TypeFile {
		copyFilename, err = utils.Xlipboard().Files()
		if err != nil {
			logger.Logger.WithError(err).Warn("failed to get path of files from clipboard")
			return
		}
		sort.Strings(copyFilename)
		return
	} else {
		err = errors.New("无法识别剪切板内容")
		return
	}
}
