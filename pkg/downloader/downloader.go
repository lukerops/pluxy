package downloader

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/lukerops/pluxy/pkg/m3u8"
)

type httpClient interface {
	Do(r *http.Request) (*http.Response, error)
}

func download(url string, client httpClient) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:105.0) Gecko/20100101 Firefox/105.0")
	req.Header.Set("Origin", "https://pluto.tv")
	req.Header.Set("Referer", "https://pluto.tv/")

	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	if response.StatusCode != 200 {
		return nil, fmt.Errorf("download failed; url: %s; status: %d", url, response.StatusCode)
	}

	return io.ReadAll(response.Body)
}

func Download(segment *m3u8.Segment, filename string) error {
	client := &http.Client{}

	segData, err := download(segment.URI, client)
	if err != nil {
		return err
	}

	if segment.Key != nil {
		keyData, err := download(segment.Key.URI, client)
		if err != nil {
			return err
		}

		// Remove o 0x
		iv := segment.Key.IV[2:]

		segData, err = decryptAesCBC(segData, keyData, []byte(iv))
		if err != nil {
			return err
		}
	}

	if err := ioutil.WriteFile(filename, segData, 0744); err != nil {
		return err
	}
	return nil
}
