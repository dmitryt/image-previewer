package resizer

import (
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
)

type Encoder interface {
	Encode(io.Writer, image.Image) error
}

type (
	JpegEncoder struct{}
	PngEncoder  struct{}
	GifEncoder  struct{}
)

func (r JpegEncoder) Encode(w io.Writer, img image.Image) error {
	return jpeg.Encode(w, img, nil)
}

func (r PngEncoder) Encode(w io.Writer, img image.Image) error {
	return png.Encode(w, img)
}

func (r GifEncoder) Encode(w io.Writer, img image.Image) error {
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
