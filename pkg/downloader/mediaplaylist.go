package downloader

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"github.com/lukerops/pluxy/pkg/bus"
	"github.com/lukerops/pluxy/pkg/commands"
	"github.com/lukerops/pluxy/pkg/m3u8"
	"github.com/lukerops/pluxy/pkg/segmentmanager"
	"github.com/rs/zerolog/log"
)

type mediaPlaylist struct {
	chRx     <-chan commands.Command
	chTx     chan<- commands.Command
	channels map[string]string

	httpDownloader *httpDownloader

	lastSeqNo uint64
	segments  []*m3u8.Segment
}

func NewMediaPlaylistWorker(httpClient *http.Client) bus.Handler {
	return &mediaPlaylist{
		httpDownloader: &httpDownloader{
			client: httpClient,
		},
        channels: make(map[string]string),
	}
}

func (mp *mediaPlaylist) Run(tx chan<- commands.Command, rx <-chan commands.Command) {
	mp.chTx = tx
	mp.chRx = rx
	go mp.run()
}

func (mp *mediaPlaylist) run() {
	logger := log.With().Str("module", "MediaPlaylist").Logger()

	logger.Info().Msg("Starting MediaPlaylist")

	for {
		cmd := <-mp.chRx

		if cmd.Cmd == commands.CommandStop {
			return
		}

		switch cmd.To {
		case commands.ToMasterPlaylist:
			masterCmd := commands.MasterPlaylist(cmd)
			mp.processMasterCommand(masterCmd)

		case commands.ToMediaPlaylist:
			mediaCmd := commands.MediaPlaylist(cmd)
			mp.processMediaCommand(mediaCmd)
		}
	}
}

func (mp *mediaPlaylist) processMasterCommand(cmd commands.MasterPlaylist) {
	if cmd.Command().IsRequest() {
		return
	}

	if !cmd.IsGetURL() {
		return
	}

	params := cmd.GetParams()
	channelID := params[0]
    playlistURL := cmd.Response

	if cmd.Command().Response == commands.ResponseFail {
		mp.chTx <- commands.NewMasterPlaylistRequest(commands.ToMediaPlaylist).GetURL(channelID).Command()
		return
	}

	mp.channels[channelID] = playlistURL
	mp.chTx <- commands.NewMediaPlaylistRequest(commands.ToMediaPlaylist).
		DownloadPlaylist(channelID, playlistURL).Command()
}

func (mp *mediaPlaylist) processMediaCommand(cmd commands.MediaPlaylist) {
	if cmd.Command().IsResponse() {
		return
	}

	params := cmd.GetParams()

	switch {
	case cmd.IsRegister():
		channelID := params[0]

		mp.channels[channelID] = ""
		mp.chTx <- commands.NewMediaPlaylistResponse(cmd, commands.ResponseOK).Command()
		mp.chTx <- commands.NewMasterPlaylistRequest(commands.ToMediaPlaylist).GetURL(channelID).Command()

	case cmd.IsDownloadPlaylist():
		channelID, playlistURL := params[0], params[1]

		rawPlaylist, err := mp.httpDownloader.DownloadFile(playlistURL)
		if err != nil {
			fmt.Println("download media playlist failed; err:", err.Error())
			break
		}

		playlist, err := m3u8.ReadManifest(string(rawPlaylist))
		if err != nil  || playlist.IsMaster() {
			fmt.Println("parse media playlist failed; err:", err.Error())
			break
		}

		if playlist.SeqNo == nil || (playlist.SeqNo != nil && *playlist.SeqNo == mp.lastSeqNo) {
			mp.chTx <- cmd.Command()
			return
		}

		for _, segment := range playlist.Segments {
			// key := ":::"
			// if segment.Key != nil {
			//     key = fmt.Sprintf("%s:::%s", segment.Key.URI, segment.Key.IV)
			// }

			mp.segments = append(mp.segments, segment)
			mp.chTx <- commands.NewMediaPlaylistRequest(commands.ToMediaPlaylist).
				DownloadSegment(channelID, len(mp.segments)-1).Command()
		}

		mp.chTx <- cmd.Command()

	case cmd.IsDownloadSegment():
		channelID, segmentIndexStr := params[0], params[1]

		segmentIndex, _ := strconv.ParseInt(segmentIndexStr, 10, 32)
		segment := mp.segments[segmentIndex]

		if segmentmanager.SegmentManager.Check(channelID, segment.URI) {
			return
		}

		pattern := regexp.MustCompile(`_ad/creative/|dai\.google\.com|Pluto_TV_OandO/.*Bumper`)
		if pattern.MatchString(segment.URI) {
			fmt.Println("Filtering Ads; url:", params)
			return
		}

		fmt.Println("Downloading segment:", segment.URI)

		segData, err := mp.httpDownloader.DownloadSegment(segment)
		if err != nil {
			fmt.Println("download segment failed; err:", err.Error())
			return
		}

		segmentmanager.SegmentManager.Add(channelID, segment.URI, segment.Duration, segData)
	}
}
