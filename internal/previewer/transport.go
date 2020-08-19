package previewer

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/dmitryt/image-previewer/internal/fetcher"
	"github.com/dmitryt/image-previewer/internal/resizer"
	"github.com/dmitryt/image-previewer/internal/utils"
	"github.com/rs/zerolog/log"
)

var ErrResize = errors.New("resize problem occurred")

type Transport struct {
	fetcher fetcher.Fetcher
	resizer *resizer.Resizer
}

func NewTransport(f fetcher.Fetcher, r *resizer.Resizer) *Transport {
	return &Transport{
		fetcher: f,
		resizer: r,
	}
}

func (t *Transport) Receive(urlParams utils.URLParams, header http.Header) (statusCode int, content string, err error) {
	pipeReader, pipeWriter := io.Pipe()
	statusCode, content, mimeType, err := t.fetcher.Fetch(urlParams.ExternalURL, header, pipeWriter)
	if err != nil {
		return
	}
	log.Debug().Msgf("File was fetched statusCode:%d err:%s", statusCode, err)
	// Resize and save to cache
	err = t.resizer.ResizeAndSave(pipeReader, urlParams, mimeType)
	if err != nil {
		statusCode = 400
		content = fmt.Sprintf("%s", ErrResize)
		if errors.Is(err, resizer.ErrUnsupportedFileType) {
			content = fmt.Sprintf("%s", err)
		}
	}

	return
}

func (t *Transport) Send(urlParams utils.URLParams, w http.ResponseWriter) (err error) {
	cacheFile, contentType, err := t.resizer.GetFile(urlParams)
	log.Debug().Msgf("Received file contentType: %s, err: %s", contentType, err)
	if err != nil {
		return
	}
	defer cacheFile.Close()
	fileInfo, err := cacheFile.Stat()
	log.Debug().Msgf("Getting file info %+v, err: %s", fileInfo, err)
	if err != nil {
		return
	}

	// Send the headers
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", strconv.FormatInt(fileInfo.Size(), 10)) // Get file size as a string

	_, err = io.Copy(w, cacheFile)

	return
}
