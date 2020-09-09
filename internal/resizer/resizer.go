package resizer

import (
	"crypto/sha512"
	"errors"
	"fmt"
	"image"
	"io"
	"os"

	"github.com/disintegration/imaging"
	"github.com/dmitryt/image-previewer/internal/cache"
	"github.com/dmitryt/image-previewer/internal/config"
	"github.com/dmitryt/image-previewer/internal/utils"
	"github.com/rs/zerolog/log"
)

type Resizer struct {
	cache cache.Cache
}

var (
	ErrInvalidURI          = errors.New("invalid URI. Expected format is: /<method>/<width>/<height>/<external url>")
	ErrRequestValidation   = errors.New("request validation error occurred")
	ErrCacheFile           = errors.New("problem with cache file occurred")
	ErrUnsupportedFileType = errors.New("file type is not supported. Supported file types: jpeg, png, gif")
)

func New(c *config.Config) (*Resizer, error) {
	ch, err := cache.New(c.CacheSize, c.CacheDir)

	return &Resizer{cache: ch}, err
}

func resize(r io.Reader, urlParams utils.URLParams) (result *image.NRGBA, err error) {
	img, _, err := image.Decode(r)
	if err != nil {
		return
	}
	result = imaging.Fill(img, urlParams.Width, urlParams.Height, imaging.Center, imaging.Lanczos)

	return
}

// Made it public just for testing purposes.
func (r *Resizer) GetCacheKey(up utils.URLParams) cache.Key {
	h := sha512.New()
	str := fmt.Sprintf("%s/%dx%d", up.ExternalURL, up.Width, up.Height)
	_, _ = io.WriteString(h, str)
	result := cache.Key([]rune(fmt.Sprintf("%x", h.Sum(nil)))[0:64])

	return result
}

func (r *Resizer) GetFile(urlParams utils.URLParams) (fd *os.File, mimeType string, err error) {
	fd, err = r.cache.GetFile(r.GetCacheKey(urlParams), os.O_RDONLY)
	if err != nil {
		return
	}
	mimeType, err = utils.GetFileMimeType(fd)

	return
}

func (r *Resizer) HasFile(urlParams utils.URLParams) bool {
	cacheKey := r.GetCacheKey(urlParams)
	_, found := r.cache.Get(cacheKey)

	return found && r.cache.HasFilePath(cacheKey)
}

func (r *Resizer) ResizeAndSave(rd io.Reader, urlParams utils.URLParams, mimeType string) (err error) {
	cacheKey := r.GetCacheKey(urlParams)
	encoder := NewEncoder(mimeType)
	if encoder == nil {
		return ErrUnsupportedFileType
	}
	_, err = r.cache.Set(cacheKey, string(cacheKey))
	if err != nil {
		return
	}
	f, err := r.cache.GetFile(cacheKey, os.O_APPEND|os.O_WRONLY)
	if err != nil {
		return
	}
	defer f.Close()
	resized, err := resize(rd, urlParams)
	log.Debug().Msgf("resizing, err: %s", err)
	if err != nil {
		return
	}
	err = encoder.Encode(f, resized.SubImage(resized.Rect))
	log.Debug().Msgf("encoding, err: %s", err)

	return
}
