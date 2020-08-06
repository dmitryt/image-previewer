package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

var cacheDir = ".cache-test"

func setup() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}

func teardown() {
	os.RemoveAll(cacheDir)
}

func prepareHandlers(server *httptest.Server) *http.ServeMux {
	api := API{Client: server.Client()}
	r := http.NewServeMux()
	r.HandleFunc("/fill/", api.ResizeHandler)
	return r
}

func checkFileInDir(t *testing.T, fileName string, expected bool) {
	_, err := os.Stat(filepath.Join(cacheDir, fileName))
	if expected {
		require.NoError(t, err)
	} else {
		require.True(t, os.IsNotExist(err))
	}
}

func makeRequest(t *testing.T, host, url string) *http.Response {
	res, err := http.Get(host + url)
	if err != nil {
		t.Fatal(t, err)
	}
	return res
}

func TestResizeCacheHandler(t *testing.T) {
	externalServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "testdata/sample.jpg")
	}))
	defer externalServer.Close()

	var externalURL = fmt.Sprintf("%s/some/file/path.jpg", strings.Replace(externalServer.URL, "http://", "", -1))

	srv := httptest.NewServer(prepareHandlers(externalServer))
	defer srv.Close()

	cacheSize, _ := strconv.Atoi(os.Getenv("CACHE_SIZE"))
	if cacheSize == 0 {
		cacheSize = 10
	}

	baseHeight := 200
	urlTemplate := "/fill/100/%d/%s"

	for i := 0; i <= cacheSize; i++ {
		url := fmt.Sprintf(urlTemplate, baseHeight+i, externalURL)
		res := makeRequest(t, srv.URL, url)
		defer res.Body.Close()

		cacheKey := GetCacheKey(url)
		require.Equal(t, http.StatusOK, res.StatusCode, "incorrect status code")
		require.Equal(t, []string{"image/jpeg"}, res.Header["Content-Type"], "incorrect Content-Type")

		checkFileInDir(t, cacheKey, true)
	}

	// Check first item. File shouldn't exist in cache folder
	url := fmt.Sprintf(urlTemplate, srv.URL, baseHeight, externalURL)
	cacheKey := GetCacheKey(url)
	checkFileInDir(t, cacheKey, false)
}
