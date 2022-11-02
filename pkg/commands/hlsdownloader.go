package commands

import (
	"fmt"
	"strconv"
	"strings"
)

type HlsDownloaderCmd struct {
	Command
}

func NewHlsDownloaderCmd(from CommandHandler) HlsDownloaderCmd {
	return HlsDownloaderCmd{
		Command{
			From: from,
			To:   HlsDownloader,
		},
	}
}

func NewHlsDownloaderCmdFrom(cmd Command) HlsDownloaderCmd {
	return HlsDownloaderCmd{cmd}
}

func NewHlsDownloaderCmdResponse(from CommandHandler, response string) HlsDownloaderCmd {
	return HlsDownloaderCmd{
		Command{
			From:       from,
			To:         HlsDownloader,
			isResponse: true,
			Response:   response,
		},
	}
}

func NewHlsDownloaderCmdResponseFrom(cmd HlsDownloaderCmd, response string) HlsDownloaderCmd {
	cmd.isResponse = true
	cmd.Response = response
	return cmd
}

func (cmd HlsDownloaderCmd) GetCommand() Command {
	return cmd.Command
}

func (cmd HlsDownloaderCmd) Register(channelID string) HlsDownloaderCmd {
	cmd.Cmd = fmt.Sprintf("REGISTER:::%s", channelID)
	return cmd
}

func (cmd HlsDownloaderCmd) IsRegister() bool {
	return strings.HasPrefix(cmd.Cmd, "REGISTER")
}

func (cmd HlsDownloaderCmd) RegisterParams() (channelID string) {
	return strings.Split(cmd.Cmd, ":::")[1]
}

func (cmd HlsDownloaderCmd) DownloadPlaylist(channelID, playlistURL string) HlsDownloaderCmd {
	cmd.Cmd = fmt.Sprintf("DOWNLOAD:::PLAYLIST:::%s:::%s", channelID, playlistURL)
	return cmd
}

func (cmd HlsDownloaderCmd) DownloadPlaylistParams() (channelID string, playlistURL string) {
	params := strings.Split(cmd.Cmd, ":::")
	return params[2], params[3]
}

func (cmd HlsDownloaderCmd) IsDownloadPlaylist() bool {
	return strings.HasPrefix(cmd.Cmd, "DOWNLOAD:::PLAYLIST")
}

func (cmd HlsDownloaderCmd) DownloadSegment(channelID string, segmentIndex int) HlsDownloaderCmd {
	cmd.Cmd = fmt.Sprintf("DOWNLOAD:::SEGMENT:::%s:::%d", channelID, segmentIndex)
	return cmd
}

func (cmd HlsDownloaderCmd) DownloadSegmentParams() (channelID string, segmentIndex int) {
	params := strings.Split(cmd.Cmd, ":::")
	index, _ := strconv.ParseInt(params[3], 10, 32)
	return params[2], int(index)
}

func (cmd HlsDownloaderCmd) IsDownloadSegment() bool {
	return strings.HasPrefix(cmd.Cmd, "DOWNLOAD:::SEGMENT")
}
