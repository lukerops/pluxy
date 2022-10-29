package downloader

import (
	"context"
	"net/http"

	"github.com/lukerops/pluxy/pkg/m3u8"
)

type httpClient interface {
	Do(r *http.Request) (*http.Response, error)
}

type Downloader interface {
	DownloadFile(context.Context, string) ([]byte, error)
	DownloadSegment(context.Context, *m3u8.Segment, string) error
}

type downloader struct {
	client httpClient
}

func NewDownloader(client httpClient) Downloader {
	return &downloader{
		client: client,
	}
}
