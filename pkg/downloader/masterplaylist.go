package downloader

import (
	"net/http"
	"net/url"
	"sort"

	"github.com/lukerops/pluxy/pkg/bus"
	"github.com/lukerops/pluxy/pkg/commands"
	"github.com/lukerops/pluxy/pkg/m3u8"
	"github.com/rs/zerolog/log"
)

type masterPlaylist struct {
	chRx         <-chan commands.Command
	chTx         chan<- commands.Command
	channels     map[string]*url.URL
	channelsHold map[string]string

	httpDownloader *httpDownloader
}

func NewMasterPlaylistWorker(httpClient *http.Client) bus.Handler {
	return &masterPlaylist{
		httpDownloader: &httpDownloader{
			client: httpClient,
		},
        channels: make(map[string]*url.URL),
        channelsHold: make(map[string]string),
	}
}

func (mp *masterPlaylist) Run(tx chan<- commands.Command, rx <-chan commands.Command) {
	mp.chTx = tx
	mp.chRx = rx

	go mp.run()
}

func (mp *masterPlaylist) run() {
	logger := log.With().Str("module", "MasterPlaylist").Logger()

	logger.Info().Msg("Starting MasterPlaylist")

	for {
		cmd := <-mp.chRx

		if cmd.Cmd == commands.CommandStop {
			return
		}

		switch cmd.To {
		case commands.ToMasterPlaylist:
			masterCmd := commands.MasterPlaylist(cmd)
			mp.processMasterCommand(masterCmd)

		case commands.ToPluto:
			plutoCmd := commands.Pluto(cmd)
			mp.processPlutoCommand(plutoCmd)

		case commands.ToMediaPlaylist:
			mediaCmd := commands.MediaPlaylist(cmd)
			mp.processMediaCommand(mediaCmd)
		}
	}
}

func (mp *masterPlaylist) processMediaCommand(cmd commands.MediaPlaylist) {
	if cmd.Command().IsRequest() {
		return
	}

	if !cmd.IsRegister() {
		return
	}

	params := cmd.GetParams()
	channelID := params[0]
	reqFrom := mp.channelsHold[channelID]

	delete(mp.channelsHold, channelID)

	if cmd.Response == commands.ResponseFail {
		mp.chTx <- commands.NewMasterPlaylistResponse(reqFrom, commands.ResponseFail).Command()
	}

	mp.channels[channelID] = nil
	mp.chTx <- commands.NewMasterPlaylistResponse(reqFrom, commands.ResponseOK).Command()
}

func (mp *masterPlaylist) processPlutoCommand(cmd commands.Pluto) {
	if cmd.Command().IsRequest() {
		return
	}

	if !cmd.IsGetURL() {
		return
	}

	params := cmd.GetParams()
	channelID := params[0]

	if cmd.Response == commands.ResponseFail {
		mp.chTx <- commands.NewPlutoRequest(commands.ToMasterPlaylist).GetURL(channelID).Command()
		return
	}

	channelURL, err := url.Parse(cmd.Response)
	if err != nil {
		mp.chTx <- commands.NewPlutoRequest(commands.ToMasterPlaylist).GetURL(channelID).Command()
		return
	}

	mp.channels[channelID] = channelURL
	mp.chTx <- commands.NewMasterPlaylistRequest(commands.ToMasterPlaylist).Download(channelID).Command()
}

func (mp *masterPlaylist) processMasterCommand(cmd commands.MasterPlaylist) {
	if cmd.Command().IsResponse() {
		return
	}

	params := cmd.GetParams()

	switch {
	case cmd.IsRegister():
		channelID := params[0]

		mp.channelsHold[channelID] = cmd.From
		mp.chTx <- commands.NewMediaPlaylistRequest(commands.ToMasterPlaylist).Register(channelID).Command()

	case cmd.IsGetURL():
		channelID := params[0]

		channelURL, ok := mp.channels[channelID]
		if !ok {
			mp.chTx <- commands.NewMasterPlaylistResponseFrom(cmd, commands.ResponseFail).Command()
			return
		}

		if channelURL == nil {
			mp.chTx <- commands.NewPlutoRequest(commands.ToMasterPlaylist).GetURL(channelID).Command()
			return
		}

		mp.chTx <- commands.NewMasterPlaylistRequest(commands.ToMasterPlaylist).Download(channelID).Command()

	case cmd.IsDownload():
		channelID := params[0]

		channelURL, ok := mp.channels[channelID]
		if !ok {
			mp.chTx <- commands.NewMasterPlaylistResponseFrom(cmd, commands.ResponseFail).Command()
			return
		}

		var retryNo int
		for retryNo = 0; retryNo < 5; retryNo += 1 {
			rawPlaylist, err := mp.httpDownloader.DownloadFile(channelURL.String())
			if err != nil {
				return
			}

			playlist, err := m3u8.ReadManifest(string(rawPlaylist))
			if err != nil || playlist.IsMedia() {
				return
			}

			// ordena as streams em ordem decrescente de
			// bandwidth, para que seja baixado a stream
			// com a maior qualidade
			sort.Slice(playlist.Streams, func(i, j int) bool {
				return playlist.Streams[i].Bandwidth > playlist.Streams[j].Bandwidth
			})

			streamURL, err := channelURL.Parse(playlist.Streams[0].URI)
			if err != nil {
				return
			}

			// envia de volta a url da atream com maior
			// qualidade
			mp.chTx <- commands.NewMasterPlaylistResponse(commands.ToMediaPlaylist, streamURL.String()).
				GetURL(channelID).Command()
			break
		}

		if retryNo == 5 {
			mp.chTx <- commands.NewMasterPlaylistResponse(commands.ToMediaPlaylist, commands.ResponseFail).
				GetURL(channelID).Command()
		}
	}
}
