package svc

import (
	"net/http"

	"github.com/lukerops/pluxy/pkg/svc/channel"
)

var (
	Channel channel.ChannelService
)

func init() {
	client := &http.Client{}

	Channel = channel.NewChannelService(client)
}
