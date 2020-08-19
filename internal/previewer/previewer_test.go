package previewer

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dmitryt/image-previewer/internal/config"
	"github.com/dmitryt/image-previewer/internal/utils"
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

func prepareHandlers(t *testing.T, cfg *config.Config, server *httptest.Server) (*Previewer, *http.ServeMux) {
	app, err := New(cfg, server.Client())
	require.NoError(t, err)
	r := http.NewServeMux()
	r.HandleFunc("/fill/", app.ResizeHandler)

	return app, r
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
	cfg := config.GetDefaultConfig()
	cfg.CacheDir = cacheDir
	externalServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "testdata/sample.jpg")
	}))
	defer externalServer.Close()

	externalURL := fmt.Sprintf("%s/some/file/path.jpg", strings.Replace(externalServer.URL, "http://", "", -1))

	app, mux := prepareHandlers(t, cfg, externalServer)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	cacheSize := cfg.CacheSize

	baseHeight := 200
	urlTemplate := "/fill/100/%d/%s"

	for i := 0; i <= cacheSize; i++ {
		url := fmt.Sprintf(urlTemplate, baseHeight+i, externalURL)
		up := utils.URLParams{ExternalURL: externalURL, Width: 100, Height: baseHeight + i}
		res := makeRequest(t, srv.URL, url)
		defer res.Body.Close()

		cacheKey := app.resizer.GetCacheKey(up)
		require.Equal(t, http.StatusOK, res.StatusCode, "incorrect status code")
		require.Equal(t, []string{"image/jpeg"}, res.Header["Content-Type"], "incorrect Content-Type")

		checkFileInDir(t, string(cacheKey), true)
	}

	// Check first item. File shouldn't exist in cache folder
	up := utils.URLParams{ExternalURL: externalURL, Width: 100, Height: baseHeight}
	cacheKey := app.resizer.GetCacheKey(up)
	checkFileInDir(t, string(cacheKey), false)
}
