package commands

import (
	"fmt"
	"strings"
)

type MasterPlaylist Command

func NewMasterPlaylistRequest(from string) MasterPlaylist {
	return MasterPlaylist{
		From: strings.ToUpper(from),
		To:   ToMasterPlaylist,
	}
}

func NewMasterPlaylistResponse(from string, response string) MasterPlaylist {
	return MasterPlaylist{
		From:       strings.ToUpper(from),
		To:         ToMasterPlaylist,
		isResponse: true,
		Response:   response,
	}
}

func NewMasterPlaylistResponseFrom(cmd MasterPlaylist, response string) MasterPlaylist {
	return MasterPlaylist{
		From:       cmd.From,
		To:         ToMasterPlaylist,
		isResponse: true,
		Response:   response,
		Cmd:        cmd.Cmd,
	}
}

func (mp MasterPlaylist) Command() Command {
	return Command(mp)
}

func (mp MasterPlaylist) GetParams() []string {
	params := strings.Split(mp.Cmd, ":::")
	switch params[0] {
	case "REGISTER":
		return params[1:]

	case "GETURL":
		return params[1:]

	case "REFRESHURL":
		return params[1:]

	case "DOWNLOAD":
		return params[1:]

	default:
		return params
	}
}

func (mp MasterPlaylist) Register(channelID string) MasterPlaylist {
	mp.Cmd = fmt.Sprintf("REGISTER:::%s", channelID)
	return mp
}

func (mp MasterPlaylist) IsRegister() bool {
	return strings.HasPrefix(mp.Cmd, "REGISTER")
}

func (mp MasterPlaylist) GetURL(channelID string) MasterPlaylist {
	mp.Cmd = fmt.Sprintf("GETURL:::%s", channelID)
	return mp
}

func (mp MasterPlaylist) IsGetURL() bool {
	return strings.HasPrefix(mp.Cmd, "GETURL")
}

func (mp MasterPlaylist) Download(channelID string) MasterPlaylist {
	mp.Cmd = fmt.Sprintf("DOWNLOAD:::%s", channelID)
	return mp
}

func (mp MasterPlaylist) IsDownload() bool {
	return strings.HasPrefix(mp.Cmd, "DOWNLOAD")
}

func (mp MasterPlaylist) RefreshURL(channelID string) MasterPlaylist {
	mp.Cmd = fmt.Sprintf("REFRESHURL:::%s", channelID)
	return mp
}

func (mp MasterPlaylist) IsRefreshURL() bool {
	return strings.HasPrefix(mp.Cmd, "REFRESHURL")
}
