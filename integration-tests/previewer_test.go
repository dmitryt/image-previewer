package main

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	diff "github.com/crhym3/imgdiff"
	"github.com/stretchr/testify/require"
)

var comparator = diff.NewBinary()

var (
	resizedImgsFolder = filepath.Join("testdata", "resized")
)

func loadImgFromHTTP(url string) (result image.Image, err error) {
	resp, err := http.Get("http://localhost:8082" + url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	result, _, err = image.Decode(resp.Body)
	return
}

func loadImgFromFile(filename string) (image.Image, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = f.Close()
	}()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}

	return img, nil
}

func TestResize(t *testing.T) {
	type test struct {
		inputURL     string
		expectedFile string
	}

	t.Run("should resize images correctly", func(t *testing.T) {
		urlString := "/fill/%d/%d/nginx/_gopher_original_1024x504.%s"
		fileNameString := "gopher_%dx%d.%s"
		sizes := [][]int{
			{50, 50},
			{200, 700},
			{256, 126},
			{333, 666},
			{500, 500},
			{1024, 252},
			{2000, 1000},
		}
		extensions := []string{"jpg", "png", "gif"}
		var tests []test
		for _, ext := range extensions {
			for _, size := range sizes {
				tests = append(tests, test{
					inputURL:     fmt.Sprintf(urlString, size[0], size[1], ext),
					expectedFile: fmt.Sprintf(fileNameString, size[0], size[1], ext),
				})
			}
		}

		for _, tc := range tests {
			expectedImg, err := loadImgFromFile(filepath.Join(resizedImgsFolder, tc.expectedFile))
			require.NoError(t, err, "Error, while reading the image from file")
			actualImg, err := loadImgFromHTTP(tc.inputURL)
			require.NoError(t, err, "Error, while fetching the image")
			_, intRes, err := comparator.Compare(expectedImg, actualImg)
			require.NoError(t, err, "Error, while comparing images")
			require.Equalf(t, 0, intRes, "Images are not equal %s", tc.expectedFile)
		}
	})
}

func TestNotSupportedFile(t *testing.T) {
	t.Run("should process not supported files correctly", func(t *testing.T) {
		resp, err := http.Get("http://localhost:8082/fill/100/100/nginx/sample.pdf")
		require.NoError(t, err, "Error, while fetching the data")
		defer resp.Body.Close()
		data, _ := ioutil.ReadAll(resp.Body)
		strData := string(data)
		require.Containsf(t, strData, "file type is not supported", "Error, while fetching the unsupported file type %s", strData)
		require.Equalf(t, resp.StatusCode, 400, "incorrect status code %d", resp.StatusCode)
	})
}

func TestFileNotFound(t *testing.T) {
	t.Run("should process not existing files correctly", func(t *testing.T) {
		resp, err := http.Get("http://localhost:8082/fill/100/100/nginx/notExist.jpg")
		require.NoError(t, err, "Error, while fetching the data")
		defer resp.Body.Close()
		data, _ := ioutil.ReadAll(resp.Body)
		strData := string(data)
		require.Containsf(t, strData, "Not Found", "Error, while fetching the not existing file %s", strData)
		require.Equalf(t, resp.StatusCode, 404, "incorrect status code %d", resp.StatusCode)
	})
}

func TestServerNotRespond(t *testing.T) {
	t.Run("should process problem with server correctly", func(t *testing.T) {
		resp, err := http.Get("http://localhost:8082/fill/100/100/invalid_server/123.jpg")
		require.NoError(t, err, "Error, while fetching the data")
		defer resp.Body.Close()
		data, _ := ioutil.ReadAll(resp.Body)
		strData := string(data)
		require.Containsf(t, strData, "no such host", "Error, while fetching the not existing file %s", strData)
		require.Equalf(t, resp.StatusCode, 502, "incorrect status code %d", resp.StatusCode)
	})
}

func TestInvalidURI(t *testing.T) {
	t.Run("should process invalid uri correctly", func(t *testing.T) {
		resp, err := http.Get("http://localhost:8082/fill/width/height/nginx/123.jpg")
		require.NoError(t, err, "Error, while fetching the data")
		defer resp.Body.Close()
		data, _ := ioutil.ReadAll(resp.Body)
		strData := string(data)
		require.Containsf(t, strData, "invalid URI", "Error, while sending the invalid URI %s", strData)
		require.Equalf(t, resp.StatusCode, 400, "incorrect status code %d", resp.StatusCode)
	})
}
