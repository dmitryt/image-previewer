package utils

import (
	"image"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/disintegration/imaging"
)

type URLParams struct {
	Method      string
	Height      int
	Width       int
	Filename    string
	ExternalURL string
	Valid       bool
}

var urlRe = regexp.MustCompile(`/fill/(\d+)/(\d+)/(.*)?`)

func ParseURL(url string) URLParams {
	if !urlRe.MatchString(url) {
		return URLParams{Valid: false}
	}
	matched := urlRe.FindAllStringSubmatch(url, -1)[0]
	// Ignore errors, non-numbers will be filtered out by regexp
	width, _ := strconv.Atoi(matched[1])
	height, _ := strconv.Atoi(matched[2])
	externalURL := matched[3]
	paths := strings.Split(externalURL, "/")

	return URLParams{
		Method:      "fill",
		Width:       width,
		Height:      height,
		Filename:    paths[len(paths)-1],
		ExternalURL: externalURL,
		Valid:       true,
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

func Atoi(str string, defaultValue int) int {
	val, err := strconv.Atoi(str)
	if err != nil {
		val = defaultValue
	}
	return val
}
