package downloader

import (
	"fmt"
	"io"
	"net/http"

	"github.com/lukerops/pluxy/pkg/m3u8"
)

type httpDownloader struct{
    client *http.Client
}

func (d *httpDownloader) DownloadFile(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
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

func (d *httpDownloader) DownloadSegment(segment *m3u8.Segment) ([]byte, error) {
	segData, err := d.DownloadFile(segment.URI)
	if err != nil {
		return nil, err
	}

	if segment.Key != nil {
		keyData, err := d.DownloadFile(segment.Key.URI)
		if err != nil {
			return nil, err
		}

		// Remove o 0x
		iv := segment.Key.IV[2:]

		segData, err = decryptAesCBC(segData, keyData, []byte(iv))
		if err != nil {
			return nil, err
		}
	}

	return segData, nil
}