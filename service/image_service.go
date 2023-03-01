package image_service

import (
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

type Image struct {
	File *os.File
	Name string
	Header []byte
	ContentType string
	NBytes int64
}

func SaveImage(files []*multipart.FileHeader, saveFileHandler func(file *multipart.FileHeader, dst string) error) error {
	for _, file := range files {
		log.Println(file.Filename)
		filename := filepath.Base(file.Filename)
		path := fmt.Sprintf("./images/%s", filename)
		if err := saveFileHandler(file, path); err != nil {
			return errors.New(fmt.Sprintf("upload file err: %s", err.Error()))
		}
	}
	return nil
}

func DownloadImage(filename string, begin int64, end int64) (*Image, error) {
	var resultImage Image
	path := fmt.Sprintf("./images/%s", filename)
	file, err := os.Open(path)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("open file err: %s", err.Error()))
	}
	fileHeader := make([]byte, 512)
	_, err = file.Read(fileHeader)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("read fileHeader err: %s", err.Error()))
	}
	resultImage.Header = fileHeader
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("file.Stat() err: %s", err.Error()))
	}
	resultImage.ContentType = http.DetectContentType(fileHeader)
	resultImage.Name = fileInfo.Name()
	if begin == 0 && end == -1 {
		file.Seek(0, 0)
		resultImage.File = file
		resultImage.NBytes = fileInfo.Size()
		return &resultImage, nil
	}

	if end == -1 {
		end = fileInfo.Size()
	}
	if begin > fileInfo.Size() || end > fileInfo.Size() {
		return nil, errors.New(fmt.Sprintf("range out of bounds for file"))
	}
	if begin >= end {
		return nil, errors.New(fmt.Sprintf("range begin cannot be bigger than range end"))
	}
	file.Seek(begin, 0)
	resultImage.File = file
	resultImage.NBytes = end - begin
	return &resultImage, nil
}
