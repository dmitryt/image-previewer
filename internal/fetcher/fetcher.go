package fetcher

import (
	"fmt"
	"context"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/dmitryt/image-previewer/internal/config"
	"github.com/dmitryt/image-previewer/internal/utils"
	"github.com/rs/zerolog/log"
)

var ErrResponseValidation = errors.New("unexpected status code >= 400")

type Fetcher interface {
	Fetch(string, http.Header, io.Writer) (int, string, string, error)
}

type HTTPFetcher struct {
	config *config.Config
	client *http.Client
}

func NewHTTPFetcher(client *http.Client, cfg *config.Config) *HTTPFetcher {
	return &HTTPFetcher{client: client, config: cfg}
}

func processData(r io.Reader, w io.Writer) (mimeType string, err error) {
	tmpFile, err := ioutil.TempFile("", "tmp")
	if err != nil {
		return
	}
	defer tmpFile.Close()
	log.Debug().Msgf("tmp file created %s", tmpFile.Name())
	if err != nil {
		return
	}

	_, err = io.Copy(tmpFile, r)
	if err != nil {
		return
	}
	log.Debug().Msg("Content was copied to tmp file")
	_, err = tmpFile.Seek(0, 0)
	if err != nil {
		return
	}
	mimeType, err = utils.GetFileMimeType(tmpFile)
	go func() {
		f, err := os.OpenFile(tmpFile.Name(), os.O_RDONLY, os.ModeAppend)
		log.Debug().Msgf("Starting writing from tmpFile %s to writer, err: %s", tmpFile.Name(), err)
		if err != nil {
			return
		}
		defer os.Remove(tmpFile.Name())
		_, err = io.Copy(w, f)
		// To handle this error need to add additional channel?
		if err != nil {
			log.Debug().Msgf("Err during copying  the file %s", err)

			return
		}
		log.Debug().Msg("Content was copied from tmp file to writer")
	}()

	return
}

func (f *HTTPFetcher) Fetch(url string, header http.Header, w io.Writer) (statusCode int, content string, mimeType string, err error) {
	statusCode = 502
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, "GET", "http://"+url, nil)
	if err != nil {
		return
	}
	req.Header = header
	resp, err := f.client.Do(req)
	if err != nil {
		content = fmt.Sprintf("%s", err)

		return
	}
	defer resp.Body.Close()

	log.Debug().Msgf("Getting the response from external server %s", resp.Status)
	if resp.StatusCode >= 400 {
		return resp.StatusCode, resp.Status, "", ErrResponseValidation
	}

	log.Debug().Msgf("Processing the data, maxFileSize: %d", f.config.MaxFileSize)
	mimeType, err = processData(io.LimitReader(resp.Body, f.config.MaxFileSize), w)
	if err != nil {
		statusCode = 500

		return
	}

	return 200, "", mimeType, nil
}
