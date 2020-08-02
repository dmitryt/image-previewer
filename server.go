package main

import (
	"io"
	"context"
	"errors"
	"os"
	"fmt"
	"strconv"
	"net/http"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/joho/godotenv"
)

var (
	ErrInvalidUri = errors.New("Invalid URI. Expected format is: /<method>/<width>/<height>/<external url>")
	ErrInvalidResponse = errors.New("External service returns invalid response")
	ErrMaxResponseSize = errors.New("Response size is greater, than accepted")
	ErrUnsupportedFileType = errors.New("File type is not supported. Supported file types: jpeg, png, gif")
)

var cache Cache

func fetch(url string, header http.Header) (*http.Response, error) {
	client := http.DefaultClient
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header = header
	return client.Do(req)
}

// Example: http://cut-service.com/fill/300/200/www.audubon.org/sites/default/files/a1_1902_16_barred-owl_sandra_rothenberg_kk.jpg
func handler(w http.ResponseWriter, r *http.Request) {
	urlParams := parseUrl("/" + r.URL.Path[1:])
	if !urlParams.valid {
		fmt.Fprintf(w, "ERRRR %s!", ErrInvalidUri)
	}
	log.Debug().Msgf("parsed url params %+v", urlParams)

	cacheDir, ok := os.LookupEnv("CACHE_DIR")
	if !ok {
		cacheDir = ".cache"
	}

	fm := FileManager{urlParams: urlParams, cacheDir: cacheDir}

	if _, ok := cache.Get(Key(fm.GetFilePath())); !ok {
		resp, err := fetch("http://" + urlParams.externalURL, r.Header)
		if err != nil {
			fmt.Fprintf(w, "ERRRR %s!", err)
		}
		defer resp.Body.Close()
		err = fm.PrepareFile(io.LimitReader(resp.Body, int64(atoi(os.Getenv("MAX_FILE_SIZE"), 5*1024*1024))))
		if err != nil {
			fmt.Fprintf(w, "ERRRR %s!", err)
		}
	}

	f, err := fm.GetFile()
	if err != nil {
		fmt.Fprintf(w, "ERRRR %s!", err)
	}
	defer f.Close()

	contentType := fm.GetFileMimeType(f)
	fileInfo, err := f.Stat()
	if err != nil {
		fmt.Fprintf(w, "ERRRR %s!", err)
	}
	//Send the headers
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", strconv.FormatInt(fileInfo.Size(), 10)) //Get file size as a string

	_, err = io.Copy(w, f)
	if err != nil {
		fmt.Fprintf(w, "ERRRR %s!", err)
	}
}

func init() {
	var val interface{}
	val, ok := os.LookupEnv("LOG_LEVEL")
	logLevel, ok := val.(zerolog.Level)
	if !ok {
		logLevel = zerolog.DebugLevel
	}
	zerolog.SetGlobalLevel(logLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: zerolog.TimeFieldFormat})
	err := godotenv.Load()
  if err != nil {
    log.Fatal().Msg("Error loading .env file")
	}
	cache = NewCache(atoi(os.Getenv("CACHE_SIZE"), 10))
}

func main() {
	http.HandleFunc("/", handler)
	if err := http.ListenAndServe("localhost:" + os.Getenv("PORT"), nil); err != nil {
    log.Fatal().Err(err).Msg("Startup failed")
	}
}
