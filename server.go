package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	ErrInvalidURI          = errors.New("invalid URI. Expected format is: /<method>/<width>/<height>/<external url>")
	ErrUnsupportedFileType = errors.New("file type is not supported. Supported file types: jpeg, png, gif")
)

var cache Cache

func fetch(url string, header http.Header) (res *http.Response, err error) {
	client := http.DefaultClient
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return
	}
	req.Header = header
	return client.Do(req)
}

func processResponse(fm *FileManager, w http.ResponseWriter) {
	f, err := fm.GetFile()
	if err != nil {
		fmt.Fprintf(w, "ERRRR %s!", err)
		return
	}
	defer f.Close()

	contentType, err := fm.GetFileMimeType(f)
	if err != nil {
		fmt.Fprintf(w, "ERRRR %s!", err)
		return
	}
	fileInfo, err := f.Stat()
	if err != nil {
		fmt.Fprintf(w, "ERRRR %s!", err)
		return
	}
	//Send the headers
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", strconv.FormatInt(fileInfo.Size(), 10)) //Get file size as a string

	_, err = io.Copy(w, f)
	if err != nil {
		fmt.Fprintf(w, "ERRRR %s!", err)
	}
}

// Example: http://cut-service.com/fill/300/200/www.audubon.org/sites/default/files/a1_1902_16_barred-owl_sandra_rothenberg_kk.jpg
func handler(w http.ResponseWriter, r *http.Request) {
	urlParams := parseURL("/" + r.URL.Path[1:])
	if !urlParams.valid {
		fmt.Fprintf(w, "ERRRR %s!", ErrInvalidURI)
		return
	}
	log.Debug().Msgf("parsed url params %+v", urlParams)

	fm := FileManager{urlParams: urlParams, cacheDir: cache.GetDir()}

	log.Debug().Msgf("Checking item %s in cache", fm.GetCacheKey())
	if _, ok := cache.Get(Key(fm.GetCacheKey())); !ok {
		log.Debug().Msg("item was not found in cache. Making request...")
		resp, err := fetch("http://"+urlParams.externalURL, r.Header)
		if err != nil {
			w.WriteHeader(502)
			fmt.Fprintf(w, "ERRRR %s!", err)
			return
		}
		if resp.StatusCode >= 400 {
			w.WriteHeader(resp.StatusCode)
			fmt.Fprint(w, resp.Status)
			return
		}
		defer resp.Body.Close()
		err = fm.PrepareFile(io.LimitReader(resp.Body, int64(atoi(os.Getenv("MAX_FILE_SIZE"), 5*1024*1024))))
		if err != nil {
			fmt.Fprintf(w, "ERRRR %s!", err)
			return
		}
	}

	processResponse(&fm, w)
}

func init() {
	var val interface{}
	var logLevel zerolog.Level
	val, ok := os.LookupEnv("LOG_LEVEL")
	if ok {
		logLevel, _ = val.(zerolog.Level)
	} else {
		logLevel = zerolog.DebugLevel
	}
	zerolog.SetGlobalLevel(logLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: zerolog.TimeFieldFormat})
	_ = godotenv.Load()
	cacheDir, ok := os.LookupEnv("CACHE_DIR")
	if !ok {
		cacheDir = ".cache"
	}
	var err error
	cache, err = NewCache(atoi(os.Getenv("CACHE_SIZE"), 10), cacheDir)
	if err != nil {
		log.Fatal().Msgf("Error, while initializing the cache: %s", err)
	}
}

func main() {
	http.HandleFunc("/", handler)
	port := atoi(os.Getenv("PORT"), 8082)
	if err := http.ListenAndServe(fmt.Sprintf("localhost:%d", port), nil); err != nil {
		log.Fatal().Err(err).Msg("Startup failed")
	}
}
