package channel

import (
	"net/http"
)

type ChannelService interface {
	GetChannelPlaylist(channelID string) (string, error)
}

type channelService struct {
	lastSeqNo                uint64
	lastFirstSegmentFilename string
}

func NewChannelService(client *http.Client) ChannelService {
	return &channelService{}
}
