package downloader

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"sort"

	"github.com/lukerops/pluxy/pkg/bus"
	"github.com/lukerops/pluxy/pkg/commands"
	"github.com/lukerops/pluxy/pkg/m3u8"
	"github.com/lukerops/pluxy/pkg/segmentmanager"
	"github.com/rs/zerolog/log"
)

type channelInfo struct {
	ID       string
	URL      *url.URL
	Provider commands.CommandHandler
}

type hlsDownloader struct {
	chRx     <-chan commands.Command
	chTx     chan<- commands.Command
	channels map[string]*channelInfo

	httpDownloader *httpDownloader

	lastSeqNo uint64
	segments  []*m3u8.Segment
}

func NewHlsDownloader(httpClient *http.Client) bus.Handler {
	return &hlsDownloader{
		channels: make(map[string]*channelInfo),
		httpDownloader: &httpDownloader{
			client: httpClient,
		},
	}
}

func (hls *hlsDownloader) Run(tx chan<- commands.Command, rx <-chan commands.Command) {
	hls.chTx = tx
	hls.chRx = rx

	go hls.run()
}

func (hls *hlsDownloader) run() {
	logger := log.With().Str("module", "HlsDownloader").Logger()

	logger.Info().Msg("Starting HlsDownloader...")

	for {
		cmd := <-hls.chRx

		if cmd.Cmd == commands.CommandStop {
			return
		}

		switch {
		case cmd.To.IsProvider():
			hls.processProvider(commands.NewProviderCmdFrom(cmd))

		case cmd.To == commands.HlsDownloader:
			hls.processCommand(commands.NewHlsDownloaderCmdFrom(cmd))
		}
	}
}

func (hls *hlsDownloader) processProvider(cmd commands.ProviderCmd) {
	if cmd.IsRequest() || !cmd.IsGetURL() {
		return
	}

	channelID := cmd.GetURLParams()
	if cmd.Response == commands.ResponseFail {
		hls.chTx <- commands.NewProviderCmd(commands.HlsDownloader, commands.PlutoProvider).
			GetURL(channelID).GetCommand()
		return
	}

	channelURL, err := url.Parse(cmd.Response)
	if err != nil {
		hls.chTx <- commands.NewProviderCmd(commands.HlsDownloader, commands.PlutoProvider).
			GetURL(channelID).GetCommand()
		return
	}

	hls.channels[channelID].URL = channelURL
	hls.chTx <- commands.NewHlsDownloaderCmd(commands.HlsDownloader).
		DownloadPlaylist(channelID, channelURL.String()).GetCommand()
}

func (hls *hlsDownloader) processCommand(cmd commands.HlsDownloaderCmd) {
	if cmd.IsResponse() {
		return
	}

	switch {
	case cmd.IsRegister():
		channelID := cmd.RegisterParams()

		hls.channels[channelID] = &channelInfo{
			ID:       channelID,
			Provider: cmd.From,
		}

		hls.chTx <- commands.NewResponseFrom(cmd.GetCommand(), commands.ResponseOK)
		hls.chTx <- commands.NewProviderCmd(commands.HlsDownloader, cmd.From).GetURL(channelID).GetCommand()

	case cmd.IsDownloadPlaylist():
		channelID, playlistURL := cmd.DownloadPlaylistParams()

		rawPlaylist, err := hls.httpDownloader.DownloadFile(playlistURL)
		if err != nil {
			return
		}

		playlist, err := m3u8.ReadManifest(string(rawPlaylist))
		if err != nil {
			return
		}

		if playlist.IsMedia() {
			if playlist.SeqNo == nil || (playlist.SeqNo != nil && *playlist.SeqNo == hls.lastSeqNo) {
				hls.chTx <- cmd.GetCommand()
				return
			}

			hls.lastSeqNo = *playlist.SeqNo
			for _, segment := range playlist.Segments {
				// filtra as propagandas
				pattern := regexp.MustCompile(`_ad/creative/|dai\.google\.com|Pluto_TV_OandO/`) //.*Bumper`)
				if pattern.MatchString(segment.URI) {
					fmt.Println("Filtering Ads; url:", segment.URI)
					continue
				}

				// não processa segmentos repetidos
				if segmentmanager.SegmentManager.Check(channelID, segment.URI) {
					continue
				}

				// key := ":::"
				// if segment.Key != nil {
				//     key = fmt.Sprintf("%s:::%s", segment.Key.URI, segment.Key.IV)
				// }

				hls.segments = append(hls.segments, segment)
				hls.chTx <- commands.NewHlsDownloaderCmd(commands.HlsDownloader).
					DownloadSegment(channelID, len(hls.segments)-1).GetCommand()
			}

			hls.chTx <- cmd.GetCommand()
			return
		}

		// ordena as streams em ordem decrescente de
		// bandwidth, para que seja baixado a stream
		// com a maior qualidade
		sort.Slice(playlist.Streams, func(i, j int) bool {
			return playlist.Streams[i].Bandwidth > playlist.Streams[j].Bandwidth
		})

		streamURL, err := hls.channels[channelID].URL.Parse(playlist.Streams[0].URI)
		if err != nil {
			return
		}

		hls.chTx <- commands.NewHlsDownloaderCmd(commands.HlsDownloader).
			DownloadPlaylist(channelID, streamURL.String()).GetCommand()

	case cmd.IsDownloadSegment():
		channelID, segmentIndex := cmd.DownloadSegmentParams()
		segment := hls.segments[segmentIndex]

		fmt.Println("Downloading segment:", segment.URI)

		segData, err := hls.httpDownloader.DownloadSegment(segment)
		if err != nil {
			fmt.Println("download segment failed; err:", err.Error())
			return
		}

		segmentmanager.SegmentManager.Add(channelID, segment.URI, segment.Duration, segData)
	}
}
