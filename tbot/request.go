package tbot

import (
	"context"
	"io"
	"net/http"
	"os"
)

func downloadFile(ctx context.Context, fileURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fileURL, nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	// Сохраняем файл на диск
	out, err := os.CreateTemp("", "*.ogg")
	if err != nil {
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}

	return out.Name(), nil
}
