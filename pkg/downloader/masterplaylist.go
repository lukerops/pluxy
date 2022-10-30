package downloader

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/lukerops/pluxy/pkg/m3u8"
	"github.com/rs/zerolog/log"
)

func (d *downloader) masterPlaylistWorker(plutoChTx, workerChTx chan<- string, plutoChRx, workerChRx <-chan string) {
	logger := log.With().Str("module", "downloader").Str("worker", "masterPlaylist").Logger()

	channels := make(map[string]*url.URL)
	internalCh := make(chan string, 2)

	for {
		select {
		case <-d.masterPlaylistStopCh:
			return

		case msg := <-plutoChRx:
			logger.Info().Str("msg", msg).Msg("chegou uma mensagem")
			params := strings.Split(msg, ":::")

			switch {
			// REGISTER:::{CHANNEL ID}
			case strings.HasPrefix(msg, "REGISTER"):
				if len(params) != 2 {
					logger.Error().Str("command", msg).Msg("invalid command")
					// plutoChTx <- fmt.Sprintf("RESPONSE:::REGISTER:::%s:::FAILED")
				}

				channels[params[1]] = nil
				workerChTx <- msg

			// RESPONSE:::GETURL:::{CHANNEL ID}:::{RESPONSE}
			case strings.HasPrefix(msg, "RESPONSE:::GETURL"):
				if len(params) != 4 {
					logger.Error().Str("command", msg).Msg("invalid command")
				}

				if params[3] == "FAILED" {
					plutoChTx <- fmt.Sprintf("GETURL:::%s", params[2])
					continue
				}

				channelURL, err := url.Parse(params[3])
				if err != nil {
					logger.Error().Stack().Err(err).Str("command", msg).
						Msg("parse channel url failed")
					plutoChTx <- fmt.Sprintf("GETURL:::%s", params[2])
					continue
				}

				channels[params[2]] = channelURL
				internalCh <- fmt.Sprintf("DOWNLOAD:::%s", params[2])
			}

		case msg := <-workerChRx:
			logger.Info().Str("msg", msg).Msg("chegou uma mensagem")
			params := strings.Split(msg, ":::")

			switch {
			// RESPONSE:::REGISTER:::{CHANNEL ID}:::{RESULT}
			case strings.HasPrefix(msg, "RESPONSE:::REGISTER"):
				if len(params) != 4 {
					logger.Error().Str("command", msg).Msg("invalid command")
				}

				if params[3] == "FAILED" {
					delete(channels, params[2])
				}

				plutoChTx <- msg

			// GETURL:::{CHANNEL ID}
			case strings.HasPrefix(msg, "GETURL"):
				if len(params) != 2 {
					logger.Error().Str("command", msg).Msg("invalid command")
				}

				channelURL, ok := channels[params[1]]
				if !ok {
					workerChTx <- fmt.Sprintf("RESPONSE:::GETURL:::%s:::FAILED", params[1])
					continue
				}

				if channelURL == nil {
					internalCh <- fmt.Sprintf("GETPLUTOURL:::%s", params[1])
					continue
				}

				internalCh <- fmt.Sprintf("DOWNLOAD:::%s", params[1])

			default:
				logger.Error().Str("command", msg).Msg("invalid command")
			}

		case msg := <-internalCh:
			logger.Info().Str("msg", msg).Msg("chegou uma mensagem")
			params := strings.Split(msg, ":::")

			switch {
			// DOWNLOAD:::{CHANNEL ID}
			case strings.HasPrefix(msg, "DOWNLOAD"):
				if len(params) != 2 {
					logger.Error().Str("command", msg).Msg("invalid command")
				}

				channelURL, ok := channels[params[1]]
				if !ok {
					workerChTx <- fmt.Sprintf("RESPONSE:::GETURL:::%s:::FAILED", params[1])
					continue
				}

				// retenta 5 vezes antes de disparar um erro
				var retryNo int
				for retryNo = 0; retryNo < 5; retryNo += 1 {
					ctx := context.Background()

					rawPlaylist, err := d.downloadFile(ctx, channelURL.String())
					if err != nil {
						logger.Error().Stack().Err(err).Str("command", msg).
							Str("channelURL", channelURL.String()).Msg("download master playlist failed")
						continue
					}

					playlist, err := m3u8.ReadMasterPlaylist(bytes.NewReader(rawPlaylist))
					if err != nil {
						logger.Error().Stack().Err(err).Str("command", msg).
							Str("channelURL", channelURL.String()).Msg("parse master playlist failed")
						continue
					}

					// ordena as streams em ordem decrescente de
					// bandwidth, para que seja baixado a stream
					// com a maior qualidade
					sort.Slice(playlist.Streams, func(i, j int) bool {
						return playlist.Streams[i].Bandwidth > playlist.Streams[j].Bandwidth
					})

					streamURL, err := channelURL.Parse(playlist.Streams[0].URI)
					if err != nil {
						logger.Error().Stack().Err(err).Str("command", msg).
							Str("channelURL", channelURL.String()).Msg("parse stream url failed")
						continue
					}

					// envia de volta a url da atream com maior
					// qualidade
					workerChTx <- fmt.Sprintf("RESPONSE:::GETURL:::%s:::%s", params[1], streamURL.String())
					break
				}

				if retryNo == 5 {
					workerChTx <- fmt.Sprintf("RESPONSE:::GETURL:::%s:::FAILED", params[1])
				}

			// GETPLUTOURL:::{CHANNEL ID}
			case strings.HasPrefix(msg, "GETPLUTOURL"):
				if len(params) != 2 {
					logger.Error().Str("command", msg).Msg("invalid command")
				}

				plutoChTx <- fmt.Sprintf("GETURL:::%s", params[1])
			}
		}
	}
}
