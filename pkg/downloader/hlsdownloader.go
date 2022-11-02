package downloader

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"time"

	"github.com/lukerops/pluxy/pkg/bus"
	"github.com/lukerops/pluxy/pkg/commands"
	"github.com/lukerops/pluxy/pkg/m3u8"
	"github.com/lukerops/pluxy/pkg/segmentmanager"
	"github.com/rs/zerolog/log"
)

type channelInfo struct {
	id         string
	uri        *url.URL
	provider   commands.CommandHandler
	lastSeqNo uint64
}

type hlsDownloader struct {
	chRx     <-chan commands.Command
	chTx     chan<- commands.Command
	channels map[string]*channelInfo

	httpDownloader *httpDownloader
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

	hls.channels[channelID].uri = channelURL
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
			id:       channelID,
			provider: cmd.From,
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
			if playlist.SeqNo != nil {
				if (*playlist.SeqNo) == hls.channels[channelID].lastSeqNo {
					bus.MessageBus.AddTimer(
						time.Duration(playlist.Segments[0].Duration/2)*time.Second, cmd.GetCommand())
					return
				}

				hls.channels[channelID].lastSeqNo = (*playlist.SeqNo)
			}

			for _, segment := range playlist.Segments {
				// filtra as propagandas
				pattern := regexp.MustCompile(`_ad/creative/|dai\.google\.com|Pluto_TV_OandO/.*Bumper`)
				if pattern.MatchString(segment.URI) {
					fmt.Println("Filtering Ads; url:", segment.URI)
					hls.chTx <- commands.NewProviderCmd(commands.HlsDownloader, hls.channels[channelID].provider).
						GetURL(channelID).GetCommand()
					return
				}

				// não processa segmentos repetidos
				if segmentmanager.SegmentManager.Check(channelID, segment.URI) {
					continue
				}

				var keyURI, keyIV string
				if segment.Key != nil {
					keyURI = segment.Key.URI
					keyIV = segment.Key.IV
				}

				hls.chTx <- commands.NewHlsDownloaderCmd(commands.HlsDownloader).
					DownloadSegment(channelID, segment.URI, keyURI, keyIV, segment.Duration).GetCommand()
			}

			bus.MessageBus.AddTimer(
				time.Duration(playlist.Segments[0].Duration/2)*time.Second, cmd.GetCommand())
			return
		}

		// ordena as streams em ordem decrescente de
		// bandwidth, para que seja baixado a stream
		// com a maior qualidade
		sort.Slice(playlist.Streams, func(i, j int) bool {
			return playlist.Streams[i].Bandwidth > playlist.Streams[j].Bandwidth
		})

		streamURL, err := hls.channels[channelID].uri.Parse(playlist.Streams[0].URI)
		if err != nil {
			return
		}

		hls.chTx <- commands.NewHlsDownloaderCmd(commands.HlsDownloader).
			DownloadPlaylist(channelID, streamURL.String()).GetCommand()

	case cmd.IsDownloadSegment():
		channelID, segmentURI, keyURI, keyIV, segmentDuration := cmd.DownloadSegmentParams()

		fmt.Println("Downloading segment:", segmentURI)

		segData, err := hls.httpDownloader.DownloadSegment(segmentURI, keyURI, keyIV)
		if err != nil {
			fmt.Println("download segment failed; err:", err.Error())
            hls.chTx <- cmd.GetCommand()
			return
		}

		segmentmanager.SegmentManager.Add(channelID, segmentURI, segmentDuration, segData)
	}
}
