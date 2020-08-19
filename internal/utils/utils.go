package utils

import (
	"errors"
	"image"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/rs/zerolog/log"
)

type URLParams struct {
	Method      string
	Height      int
	Width       int
	Filename    string
	ExternalURL string
	Error       error
}

var ErrURLPatternMatching = errors.New("URL doesn't match the pattern")

var urlRe = regexp.MustCompile(`/fill/(\d+)/(\d+)/(.*)?`)

func ParseURL(url string) URLParams {
	if !urlRe.MatchString(url) {
		return URLParams{Error: ErrURLPatternMatching}
	}
	matched := urlRe.FindAllStringSubmatch(url, -1)[0]
	width, errWidth := strconv.Atoi(matched[1])
	height, errHeight := strconv.Atoi(matched[2])
	externalURL := matched[3]
	paths := strings.Split(externalURL, "/")

	err := errWidth
	if err == nil {
		err = errHeight
	}

	return URLParams{
		Method:      "fill",
		Width:       width,
		Height:      height,
		Filename:    paths[len(paths)-1],
		ExternalURL: externalURL,
		Error:       err,
	}
}

func Resize(r io.Reader, urlParams URLParams) (result *image.NRGBA, err error) {
	img, _, err := image.Decode(r)
	if err != nil {
		return
	}
	result = imaging.Fill(img, urlParams.Width, urlParams.Height, imaging.Center, imaging.Lanczos)

	return
}

func GetFileMimeType(f *os.File) (result string, err error) {
	fileHeader := make([]byte, 512)
	_, err = f.Read(fileHeader)
	if err != nil {
		log.Debug().Msgf("reading file header error %s", err)

		return
	}
	_, err = f.Seek(0, 0)
	if err != nil {
		log.Debug().Msgf("seeking the file error %s", err)

		return
	}
	// Get content type of file
	result = http.DetectContentType(fileHeader)

	return
}
