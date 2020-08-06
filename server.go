package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	. "github.com/dmitryt/image-previewer/internal/cache"
	. "github.com/dmitryt/image-previewer/internal/file_manager"
	utils "github.com/dmitryt/image-previewer/internal/utils"
)

type DummyResponse struct {
	OK bool
}

var (
	ErrInvalidURI        = errors.New("invalid URI. Expected format is: /<method>/<width>/<height>/<external url>")
	ErrRequestValidation = errors.New("request validation error occurred")
	ErrCacheFile         = errors.New("problem with cache file occurred")
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

func processResponse(fm *FileManager, w http.ResponseWriter, cacheFile *os.File) {
	contentType, err := fm.GetFileMimeType(cacheFile)
	if err != nil {
		fmt.Fprintf(w, "ERROR %s!", err)
		return
	}
	fileInfo, err := cacheFile.Stat()
	if err != nil {
		fmt.Fprintf(w, "ERROR %s!", err)
		return
	}
	//Send the headers
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", strconv.FormatInt(fileInfo.Size(), 10)) //Get file size as a string

	_, err = io.Copy(w, cacheFile)
	if err != nil {
		fmt.Fprintf(w, "ERROR %s!", err)
	}
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	dummyResponse := DummyResponse{true}

	content, err := json.Marshal(dummyResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(content)
}

func processRequest(w http.ResponseWriter, fm *FileManager, header http.Header) {
	cacheKey := cache.GetKey(fm.UrlParams)
	log.Debug().Msgf("Checking item %s in cache", cacheKey)
	if value, ok := cache.Get(cacheKey); ok {
		// Refreshing the value in cache
		_, err := cache.Set(cacheKey, value)
		if err != nil {
			log.Error().Msgf("%s %s", ErrCacheFile, err)
		}
		return
	}
	resp, err := fetch("http://"+fm.UrlParams.ExternalURL, header)
	if err != nil {
		log.Error().Msgf("%s", err)
		w.WriteHeader(502)
		fmt.Fprintf(w, "ERROR %s!", err)
		return
	}
	if resp.StatusCode >= 400 {
		log.Error().Msgf("%s %s", ErrRequestValidation, resp.Status)
		w.WriteHeader(resp.StatusCode)
		fmt.Fprint(w, resp.Status)
		return
	}
	defer resp.Body.Close()

	_, err = cache.Set(cacheKey, cache.GetFilePath(fm.UrlParams))
	if err != nil {
		log.Error().Msgf("%s %s", ErrCacheFile, err)
	}
	f, err := cache.GetFile(fm.UrlParams)
	if err != nil {
		log.Error().Msgf("%s %s", ErrCacheFile, err)
		w.WriteHeader(400)
		fmt.Fprintf(w, "ERROR %s!", ErrCacheFile)
		return
	}
	defer f.Close()
	err = fm.PrepareFile(io.LimitReader(resp.Body, int64(utils.Atoi(os.Getenv("MAX_FILE_SIZE"), 5*1024*1024))), f)
	if err != nil {
		log.Error().Msgf("%s", err)
		w.WriteHeader(400)
		fmt.Fprintf(w, "ERROR %s!", err)
		return
	}
}

// Example: http://cut-service.com/fill/300/200/www.audubon.org/sites/default/files/a1_1902_16_barred-owl_sandra_rothenberg_kk.jpg
func mainHandler(w http.ResponseWriter, r *http.Request) {
	urlParams := utils.ParseURL("/" + r.URL.Path[1:])
	if !urlParams.Valid {
		log.Error().Msgf("%s %+v", ErrInvalidURI, urlParams)
		w.WriteHeader(400)
		fmt.Fprintf(w, "ERROR %s!", ErrInvalidURI)
		return
	}
	log.Debug().Msgf("parsed url params %+v", urlParams)

	fm := FileManager{UrlParams: urlParams}

	processRequest(w, &fm, r.Header)
	f, err := cache.GetFile(fm.UrlParams)
	if err != nil {
		log.Error().Msgf("%s %s", ErrCacheFile, err)
		w.WriteHeader(400)
		fmt.Fprintf(w, "ERROR %s!", ErrCacheFile)
		return
	}
	defer f.Close()
	processResponse(&fm, w, f)
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
	cache, err = NewCache(utils.Atoi(os.Getenv("CACHE_SIZE"), 10), cacheDir)
	if err != nil {
		log.Fatal().Msgf("Error, while initializing the cache: %s", err)
	}
}

func main() {
	http.HandleFunc("/health_check", healthCheckHandler)
	http.HandleFunc("/fill/", mainHandler)

	address := fmt.Sprintf("0.0.0.0:%d", utils.Atoi(os.Getenv("PORT"), 8082))

	log.Info().Msgf("Server starting at %s", address)
	if err := http.ListenAndServe(address, nil); err != nil {
		log.Fatal().Err(err).Msg("Startup failed")
	}
}
