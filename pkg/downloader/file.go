package downloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func (d *downloader) DownloadFile(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:105.0) Gecko/20100101 Firefox/105.0")
	req.Header.Set("Origin", "https://pluto.tv")
	req.Header.Set("Referer", "https://pluto.tv/")

	response, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	if response.StatusCode != 200 {
		return nil, fmt.Errorf("download failed; url: %s; status: %d", url, response.StatusCode)
	}

	return io.ReadAll(response.Body)
}
