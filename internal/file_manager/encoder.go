package file_manager

import (
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
)

type Encoder interface {
	encode(io.Writer, image.Image) error
}

type JpegEncoder struct{}
type PngEncoder struct{}
type GifEncoder struct{}

func (r JpegEncoder) encode(w io.Writer, img image.Image) error {
	return jpeg.Encode(w, img, nil)
}

func (r PngEncoder) encode(w io.Writer, img image.Image) error {
	return png.Encode(w, img)
}

func (r GifEncoder) encode(w io.Writer, img image.Image) error {
	return gif.Encode(w, img, nil)
}

func NewEncoder(contentType string) Encoder {
	switch contentType {
	case "image/jpeg":
		return JpegEncoder{}
	case "image/png":
		return PngEncoder{}
	case "image/gif":
		return GifEncoder{}
	default:
		return nil
	}
}
