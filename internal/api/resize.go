package api

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/rs/zerolog/log"

	cachePkg "github.com/dmitryt/image-previewer/internal/cache"
	fmPkg "github.com/dmitryt/image-previewer/internal/file_manager"
	utils "github.com/dmitryt/image-previewer/internal/utils"
)

var (
	ErrInvalidURI        = errors.New("invalid URI. Expected format is: /<method>/<width>/<height>/<external url>")
	ErrRequestValidation = errors.New("request validation error occurred")
	ErrCacheFile         = errors.New("problem with cache file occurred")
)

var cache cachePkg.Cache

func (api *API) fetch(url string, header http.Header) (res *http.Response, err error) {
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return
	}
	req.Header = header
	return api.Client.Do(req)
}

func (api *API) processResponse(fm *fmPkg.FileManager, w http.ResponseWriter, cacheFile *os.File) {
	contentType, err := fm.GetFileMimeType(cacheFile)
	log.Debug().Msgf("Detecting the contentType %s", contentType)
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

func (api *API) processRequest(fm *fmPkg.FileManager, header http.Header) (statusCode int, response string) {
	cacheKey := cache.GetKey(fm.URLParams)
	log.Debug().Msgf("Checking item %s in cache", cacheKey)
	if _, ok := cache.Get(cacheKey); ok {
		// Doing nothing in this case
		return
	}
	resp, err := api.fetch("http://"+fm.URLParams.ExternalURL, header)
	if err != nil {
		log.Error().Msgf("%s", err)
		return 502, fmt.Sprintf("%s", err)
	}
	log.Debug().Msgf("Getting the response from external server %s", resp.Status)
	if resp.StatusCode >= 400 {
		log.Error().Msgf("%s: %s", ErrRequestValidation, resp.Status)
		return resp.StatusCode, resp.Status
	}
	defer resp.Body.Close()

	_, err = cache.Set(cacheKey, string(cacheKey))
	if err != nil {
		log.Error().Msgf("%s %s", ErrCacheFile, err)
	}
	f, err := cache.GetFile(fm.URLParams)
	if err != nil {
		log.Error().Msgf("%s %s", ErrCacheFile, err)
		return 400, fmt.Sprintf("%s", ErrCacheFile)
	}
	defer f.Close()
	err = fm.PrepareFile(io.LimitReader(resp.Body, int64(utils.Atoi(os.Getenv("MAX_FILE_SIZE"), 5*1024*1024))), f)
	if err != nil {
		log.Error().Msgf("%s", err)
		return resp.StatusCode, fmt.Sprintf("%s", err)
	}
	return 200, ""
}

// For testing purposes.
func GetCacheKey(url string) string {
	urlParams := utils.ParseURL(url)
	fm := fmPkg.FileManager{URLParams: urlParams}
	return string(cache.GetKey(fm.URLParams))
}

func init() {
	cacheDir, ok := os.LookupEnv("CACHE_DIR")
	if !ok {
		cacheDir = ".cache"
	}
	var err error
	cache, err = cachePkg.NewCache(utils.Atoi(os.Getenv("CACHE_SIZE"), 10), cacheDir)
	if err != nil {
		log.Fatal().Msgf("Error, while initializing the cache: %s", err)
	}
}

// Example: http://cut-service.com/fill/300/200/www.audubon.org/sites/default/files/a1_1902_16_barred-owl_sandra_rothenberg_kk.jpg
func (api *API) ResizeHandler(w http.ResponseWriter, r *http.Request) {
	urlParams := utils.ParseURL("/" + r.URL.Path[1:])
	if !urlParams.Valid {
		log.Error().Msgf("%s %+v", ErrInvalidURI, urlParams)
		w.WriteHeader(400)
		fmt.Fprintf(w, "ERROR %s!", ErrInvalidURI)
		return
	}
	log.Debug().Msgf("parsed url params %+v", urlParams)

	fm := fmPkg.FileManager{URLParams: urlParams}

	statusCode, errStr := api.processRequest(&fm, r.Header)
	if statusCode >= 400 {
		w.WriteHeader(statusCode)
		fmt.Fprintf(w, "ERROR %s!", errStr)
		return
	}
	f, err := cache.GetFile2(fm.URLParams)
	if err != nil {
		log.Error().Msgf("%s %s", ErrCacheFile, err)
		w.WriteHeader(400)
		fmt.Fprintf(w, "ERROR %s!", ErrCacheFile)
		return
	}
	defer f.Close()
	api.processResponse(&fm, w, f)
}
