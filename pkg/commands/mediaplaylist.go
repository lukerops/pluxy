package commands

import (
	"fmt"
	"strings"
)

type MediaPlaylist Command

func NewMediaPlaylistRequest(from string) MediaPlaylist {
	return MediaPlaylist{
		From: strings.ToUpper(from),
		To:   ToMediaPlaylist,
	}
}

func NewMediaPlaylistResponse(request MediaPlaylist, response string) MediaPlaylist {
	return MediaPlaylist{
		From:       request.From,
		To:         ToMediaPlaylist,
		isResponse: true,
		Response:   response,
		Cmd:        request.Cmd,
	}
}

func (mp MediaPlaylist) Command() Command {
	return Command(mp)
}

func (mp MediaPlaylist) GetParams() []string {
	params := strings.Split(mp.Cmd, ":::")

	switch params[0] {
	case "REGISTER":
		return params[1:]

	case "DOWNLOAD":
		return params[2:]

	default:
		return params
	}
}

func (mp MediaPlaylist) Register(channelID string) MediaPlaylist {
	mp.Cmd = fmt.Sprintf("REGISTER:::%s", channelID)
	return mp
}

func (mp MediaPlaylist) IsRegister() bool {
	return strings.HasPrefix(mp.Cmd, "REGISTER")
}

func (mp MediaPlaylist) DownloadPlaylist(channelID, playlistURL string) MediaPlaylist {
	mp.Cmd = fmt.Sprintf("DOWNLOAD:::PLAYLIST:::%s:::%s", channelID, playlistURL)
	return mp
}

func (mp MediaPlaylist) IsDownloadPlaylist() bool {
	return strings.HasPrefix(mp.Cmd, "DOWNLOAD:::PLAYLIST")
}

func (mp MediaPlaylist) DownloadSegment(channelID string, segmentIndex int) MediaPlaylist {
	mp.Cmd = fmt.Sprintf("DOWNLOAD:::SEGMENT:::%s:::%d", channelID, segmentIndex)
	return mp
}

func (mp MediaPlaylist) IsDownloadSegment() bool {
	return strings.HasPrefix(mp.Cmd, "DOWNLOAD:::SEGMENT")
}
