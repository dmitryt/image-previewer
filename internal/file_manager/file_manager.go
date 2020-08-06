package file_manager

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	utils "github.com/dmitryt/image-previewer/internal/utils"

	"github.com/rs/zerolog/log"
)

var (
	ErrUnsupportedFileType = errors.New("file type is not supported. Supported file types: jpeg, png, gif")
)

type FileManager struct {
	UrlParams utils.URLParams
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

func (fm FileManager) PrepareFile(r io.Reader, w io.Writer) (err error) {
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
	resized, err := utils.Resize(tmpFile, fm.UrlParams)

	err = encoder.encode(w, resized.SubImage(resized.Rect))
	return
}