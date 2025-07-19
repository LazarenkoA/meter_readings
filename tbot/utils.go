package tbot

import (
	"github.com/pkg/errors"
	"io"
	"net/http"
	"os"
)

func castMap[T any](data map[string]interface{}) map[string]T {
	result := make(map[string]T, len(data))

	for k, v := range data {
		if vv, ok := v.(T); ok {
			result[k] = vv
		}
	}

	return result
}

func fileExist(filePath string) bool {
	if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
		return false
	}

	return true
}

func (t *tBot) downloadFile(fileUrl, baseFileName string) (string, error) {
	resp, err := http.Get(fileUrl)
	if err != nil {
		return "", errors.Wrap(err, "http get error")
	}
	defer resp.Body.Close()

	// Сохраняем файл локально
	outFile, err := os.CreateTemp("", "*"+baseFileName)
	if err != nil {
		return "", err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return "", err
	}

	return outFile.Name(), nil
}
