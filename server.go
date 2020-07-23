package main

import (
	"context"
	"errors"
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/disintegration/imaging"
)

var cacheDir = ".cache"

var (
	ErrInvalidUri = errors.New("Invalid URI. Expected format is: /<method>/<width>/<height>/<external url>")
)

type UrlParams struct {
	method      string
	height      int
	width       int
	filename    string
	externalUrl string
}

func atoi(val string) int {
	result, err := strconv.Atoi(val)
	if err != nil {
		return 0
	}
	return result
}

func parseUrl(url string) (UrlParams, error) {
	paths := strings.Split(url, "/")
	if len(paths) < 4 {
		return UrlParams{}, ErrInvalidUri
	}
	return UrlParams{
		method:      paths[0],
		width:       atoi(paths[1]),
		height:      atoi(paths[2]),
		filename:    paths[len(paths)-1],
		externalUrl: strings.Join(paths[3:], "/"),
	}, nil
}

// Example: http://cut-service.com/fill/300/200/www.audubon.org/sites/default/files/a1_1902_16_barred-owl_sandra_rothenberg_kk.jpg
func handler(w http.ResponseWriter, r *http.Request) {
	urlParams, err := parseUrl(r.URL.Path[1:])
	if err != nil {
		fmt.Fprintf(w, "ERRRR %s!", err)
	}
	log.Println("[DEBUG] parsed url params", urlParams)
	client := http.DefaultClient
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, "GET", "http://"+urlParams.externalUrl, nil)
	if err != nil {
		fmt.Fprintf(w, "ERRRR %s!", err)
		return
	}
	req.Header = r.Header
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(w, "ERRRR %s!", err)
		return
	}
	log.Println("[DEBUG] fetched image")
	defer resp.Body.Close()

	// Prepare dirs
	baseDirPath := path.Join(cacheDir, urlParams.externalUrl, fmt.Sprintf("%sx%s", urlParams.width, urlParams.height))
	err = os.MkdirAll(baseDirPath, 0755)
	if err != nil {
		fmt.Fprintf(w, "ERRRR %s!", err)
		return
	}
	log.Println("[DEBUG] created directories", baseDirPath)

	// Create file
	filePath := path.Join(baseDirPath, urlParams.filename)
	f, err := os.Create(filePath)
	if err != nil {
		fmt.Fprintf(w, "ERRRR %s!", err)
		return
	}
	log.Println("[DEBUG] created file", filePath)
	defer f.Close()

	// Copy data to file
	decoded, _, err := image.Decode(resp.Body)
	if err != nil {
		fmt.Fprintf(w, "ERRRR %s!", err)
		return
	}
	resized, err := resize(&decoded, urlParams.width, urlParams.height)
	_, err = f.Write(resized)
	if err != nil {
		fmt.Fprintf(w, "ERRRR %s!", err)
		return
	}

	// Retunr to client
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(w, "ERRRR %s!", err)
		return
	}
	log.Println("[DEBUG] return to client", filePath)
	_, err = w.Write(content)
	if err != nil {
		fmt.Fprintf(w, "ERRRR %s!", err)
		return
	}
}

func resize(src *image.Image, width, height int) ([]byte, error) {
	buf := new(bytes.Buffer)
	imaging := imaging.Resize(*src, width, height, imaging.Lanczos)
	err := jpeg.Encode(buf, imaging, nil)
	if err != nil {
		return nil, err
	} else {
		return buf.Bytes(), nil
	}
}

func main() {
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
