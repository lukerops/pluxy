package downloader

import (
	"net/http"
)

type Downloader interface {
	Run(plutoChTx chan<- string, plutoChRx <-chan string)
}

type downloader struct {
	client               *http.Client
	masterPlaylistStopCh chan struct{}
	mediaPlaylistStopCh  chan struct{}
}

func NewDownloader() Downloader {
	return &downloader{
		client: &http.Client{},
	}
}
