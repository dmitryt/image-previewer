package main

import (
	"crypto/md5"
	"io"
	"os"
	"io/ioutil"
	"path"
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"
)

type FileManager struct {
	cacheDir string
	urlParams URLParams
}

type FileInfo struct {
	fileContentType string;
	fileSize int64;
}

func (fm FileManager) GetDirPath() string {
	width := fm.urlParams.width
	height := fm.urlParams.height
	h := md5.New()
	io.WriteString(h, fmt.Sprintf("%s/%dx%d", fm.urlParams.externalURL, width, height))
	return path.Join(fm.cacheDir, fmt.Sprintf("%x", h.Sum(nil)))
}

func (fm FileManager) GetFilePath() string {
	return path.Join(fm.GetDirPath(), fm.urlParams.filename)
}

func (fm FileManager) GetFile() (*os.File, error) {
	return os.Open(fm.GetFilePath())
}

func (fm FileManager) GetFileMimeType(f *os.File) string {
	fileHeader := make([]byte, 512)
	f.Read(fileHeader)
	f.Seek(0, 0)
	//Get content type of file
	return http.DetectContentType(fileHeader)
}

func (fm FileManager) PrepareFile(r io.Reader) (err error) {
	// Init Tmp File
	tmpFile, err := ioutil.TempFile("", "tmp")
	if err != nil {
		return
	}
	io.Copy(tmpFile, r)
	tmpFile.Seek(0, 0)
	log.Debug().Msgf("tmp file created %s", tmpFile.Name())
	defer os.Remove(tmpFile.Name())

	// Prepare dirs
	baseDirPath := fm.GetDirPath()
	err = os.MkdirAll(baseDirPath, 0755)
	if err != nil {
		return
	}
	log.Debug().Msgf("created directories %s", baseDirPath)

	mimeType := fm.GetFileMimeType(tmpFile)
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