package main

import (
	"crypto/sha512"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"

	"github.com/rs/zerolog/log"
)

type FileManager struct {
	cacheDir  string
	urlParams URLParams
}

func (fm FileManager) GetCacheKey() string {
	width := fm.urlParams.width
	height := fm.urlParams.height
	h := sha512.New()
	_, _ = io.WriteString(h, fmt.Sprintf("%s/%dx%d", fm.urlParams.externalURL, width, height))
	return string([]rune(fmt.Sprintf("%x", h.Sum(nil)))[0:64])
}

func (fm FileManager) GetFilePath() string {
	return path.Join(fm.cacheDir, fm.GetCacheKey())
}

func (fm FileManager) GetFile() (*os.File, error) {
	return os.Open(fm.GetFilePath())
}

func (fm FileManager) GetFileMimeType(f *os.File) (result string, err error) {
	fileHeader := make([]byte, 512)
	_, err = f.Read(fileHeader)
	if err != nil {
		return
	}
	_, err = f.Seek(0, 0)
	if err != nil {
		return
	}
	//Get content type of file
	result = http.DetectContentType(fileHeader)
	return
}

func (fm FileManager) PrepareFile(r io.Reader) (err error) {
	// Init Tmp File
	tmpFile, err := ioutil.TempFile("", "tmp")
	if err != nil {
		return
	}
	_, err = io.Copy(tmpFile, r)
	if err != nil {
		return
	}
	_, err = tmpFile.Seek(0, 0)
	if err != nil {
		return
	}
	log.Debug().Msgf("tmp file created %s", tmpFile.Name())
	defer os.Remove(tmpFile.Name())

	mimeType, err := fm.GetFileMimeType(tmpFile)
	if err != nil {
		return
	}
	log.Debug().Msgf("found mimeType: %s", mimeType)
	encoder := NewEncoder(mimeType)
	if encoder == nil {
		return ErrUnsupportedFileType
	}
	resized, err := resize(tmpFile, fm.urlParams)

	// Create file
	f, err := os.Create(fm.GetFilePath())
	if err != nil {
		return
	}
	log.Debug().Msgf("created file %s", fm.GetFilePath())
	defer f.Close()
	err = encoder.encode(f, resized.SubImage(resized.Rect))
	return
}
