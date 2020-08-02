package main

import (
	"io"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"image"

	"github.com/disintegration/imaging"
)

type URLParams struct {
	method      string
	height      int
	width       int
	filename    string
	externalURL string
	valid bool
}

var urlRe = regexp.MustCompile(`/fill/(\d+)/(\d+)/(.*)?`)

func parseUrl(url string) URLParams {
	if !urlRe.MatchString(url) {
		return URLParams{valid: false}
	}
	matched := urlRe.FindAllStringSubmatch(url, -1)[0]
	// Ignore errors, non-numbers will be filtered out by regexp
	width, _ := strconv.Atoi(matched[1])
	height, _ := strconv.Atoi(matched[2])
	externalURL := matched[3]
	paths := strings.Split(externalURL, "/")

	return URLParams{
		method:      matched[0],
		width:       width,
		height:      height,
		filename:    paths[len(paths)-1],
		externalURL: externalURL,
		valid: true,
	}
}

func stringifyURL(u URLParams) string {
	return fmt.Sprintf("/%d/%d/%s", u.width, u.height, u.externalURL)
}

func resize(r io.Reader, urlParams URLParams) (result *image.NRGBA, err error) {
	img, _, err := image.Decode(r)
	if err != nil {
		return
	}
	result = imaging.Fill(img, urlParams.width, urlParams.height, imaging.Center, imaging.Lanczos)
	return
}

func atoi(str string, defaultValue int) int {
	val, err := strconv.Atoi(str)
	if err != nil {
		val = defaultValue
	}
	return val
}
