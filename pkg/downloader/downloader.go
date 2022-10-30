package downloader

import (
	"context"
	"net/http"
)

type Downloader interface {
	Run(ctx context.Context)
}

type downloader struct {
	client     *http.Client
	channelURL string
	channelID  string
}

func NewDownloader(channelID, channelURL string) Downloader {
	return &downloader{
		client:     &http.Client{},
		channelID:  channelID,
		channelURL: channelURL,
	}
}
